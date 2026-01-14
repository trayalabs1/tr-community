package traya

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/account/authentication"
	"github.com/Southclaws/storyden/app/resources/account/email"
	"github.com/Southclaws/storyden/app/resources/channel"
	"github.com/Southclaws/storyden/app/resources/channel_membership"
	"github.com/Southclaws/storyden/app/services/account/register"
	"github.com/Southclaws/storyden/internal/config"
)

var (
	ErrTrayaAPIFailed      = fault.New("traya API request failed", ftag.With(ftag.Internal))
	ErrInvalidToken        = fault.New("invalid token", ftag.With(ftag.PermissionDenied))
	ErrAccountNotFound     = fault.New("account not found", ftag.With(ftag.NotFound))
	ErrMissingAPIServerURL = fault.New("TRAYA_API_SERVER_URL not configured", ftag.With(ftag.Internal))
)

var service = authentication.ServiceTraya

type Provider struct {
	logger         *slog.Logger
	cfg            config.Config
	authRepo       authentication.Repository
	accountQuery   *account_querier.Querier
	accountWriter  *account_writer.Writer
	emailRepo      *email.Repository
	register       *register.Registrar
	channelRepo    *channel.Repository
	membershipRepo *channel_membership.Repository
	httpClient     *http.Client
}

func New(
	logger *slog.Logger,
	cfg config.Config,
	authRepo authentication.Repository,
	accountQuery *account_querier.Querier,
	accountWriter *account_writer.Writer,
	emailRepo *email.Repository,
	register *register.Registrar,
	channelRepo *channel.Repository,
	membershipRepo *channel_membership.Repository,
) *Provider {
	return &Provider{
		logger:         logger,
		cfg:            cfg,
		authRepo:       authRepo,
		accountQuery:   accountQuery,
		accountWriter:  accountWriter,
		emailRepo:      emailRepo,
		register:       register,
		channelRepo:    channelRepo,
		membershipRepo: membershipRepo,
		httpClient:     &http.Client{},
	}
}

func (p *Provider) Service() authentication.Service { return service }
func (p *Provider) Token() authentication.TokenType { return authentication.TokenTypeTraya }

type TrayaUserResponse struct {
	User struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		PhoneNumber string  `json:"phone_number"`
		FirstName   string  `json:"first_name"`
		LastName    *string `json:"last_name"`
		Gender      string  `json:"gender"`
		CreatedAt   string  `json:"createdAt"`
	} `json:"user"`
	Case struct {
		ID                      string  `json:"id"`
		LatestOrderCount        int     `json:"latest_order_count"`
		LatestOrderDate         string  `json:"latest_order_date"`
		LatestOrderDisplayID    string  `json:"latest_order_display_id"`
		LatestOrderID           string  `json:"latest_order_id"`
		LatestOrderStatus       string  `json:"latest_order_status"`
		LatestOrderDeliveryDate *string `json:"latest_order_delivery_date"`
		FormStatus              string  `json:"form_status"`
	} `json:"case"`
	ChatURL                string `json:"chatUrl"`
	TotalKitCount          int    `json:"totalKitCount"`
	RunningMonthForHairKit int    `json:"runningMonthForHairKit"`
	CustomerSlug           struct {
		SlugName any `json:"slugName"`
	} `json:"customerSlug"`
}

type cohortChannelRule struct {
	orderCount  int
	channelSlug string
	gender      string
}

var cohortChannelRules = []cohortChannelRule{
	{1, "month-1-warriors", "male"},
	{2, "month-2-titans", "male"},
	{3, "month-3-pioneers", "male"},
	{4, "month-4-mavericks", "male"},
	{5, "month-5-masters", "male"},
	{6, "month-6-8-veterans", "male"},
	{7, "month-6-8-veterans", "male"},
	{8, "month-6-8-veterans", "male"},
	{9, "month-8-plus-legends", "male"},

	{1, "month-1-heroines", "female"},
	{2, "month-2-divas", "female"},
	{3, "month-3-icons", "female"},
	{4, "month-4-anchors", "female"},
	{5, "month-5-elites", "female"},
	{6, "month-6-8-champions", "female"},
	{7, "month-6-8-champions", "female"},
	{8, "month-6-8-champions", "female"},
	{9, "month-8-plus-queens", "female"},
}

var topicChannelsByGender = map[string][]string{
	"male": {
		"general",
		"stress-sleep-nutrition",
		"digestion-metabolism-gut-health",
		"dandruff-hair-health",
	},
	"female": {
		"general",
		"hormones-pcos",
		"stress-sleep-nutrition-female",
		"digestion-metabolism-gut-female",
		"dandruff-hair-health-female",
	},
}

func getAllManagedChannelSlugs() map[string]bool {
	slugs := make(map[string]bool)
	for _, rule := range cohortChannelRules {
		slugs[rule.channelSlug] = true
	}
	for _, channels := range topicChannelsByGender {
		for _, slug := range channels {
			slugs[slug] = true
		}
	}
	return slugs
}

func (p *Provider) AuthenticateWithToken(ctx context.Context, token string) (*account.Account, error) {
	if p.cfg.TrayaAPIServerURL.String() == "" {
		return nil, fault.Wrap(ErrMissingAPIServerURL, fctx.With(ctx))
	}

	userData, err := p.fetchTrayaUserData(ctx, token)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	emailAddress, err := mail.ParseAddress(userData.User.Email)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), ftag.With(ftag.InvalidArgument))
	}

	orderCount := userData.Case.LatestOrderCount
	if orderCount == 0 {
		orderCount = userData.TotalKitCount
	}

	acc, err := p.getOrCreateAccount(ctx, userData.User.ID, *emailAddress, userData.User.FirstName, userData.User.LastName, userData.User.PhoneNumber)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	// Update account metadata with case_id from Traya API (only if not already set)
	if acc.Metadata == nil || acc.Metadata["case_id"] == nil {
		metadata := acc.Metadata
		if metadata == nil {
			metadata = make(map[string]any)
		}
		metadata["case_id"] = userData.Case.ID

		accWithMetadata, err := p.accountWriter.Update(ctx, acc.ID, account_writer.SetMetadata(metadata))
		if err != nil {
			p.logger.Warn("failed to update account metadata",
				slog.String("account_id", acc.ID.String()),
				slog.String("case_id", userData.Case.ID),
				slog.String("error", err.Error()))
			// Continue with authentication even if metadata update fails
		} else {
			acc = &accWithMetadata.Account
		}
	}

	if err := p.ensureChannelMemberships(ctx, acc.ID, userData.User.Gender, orderCount, userData.Case.LatestOrderDate); err != nil {
		p.logger.Warn("failed to ensure channel memberships",
			slog.String("account_id", acc.ID.String()),
			slog.Int("order_count", orderCount),
			slog.String("gender", userData.User.Gender),
			slog.String("error", err.Error()))
	}

	return acc, nil
}

func (p *Provider) fetchTrayaUserData(ctx context.Context, token string) (*TrayaUserResponse, error) {
	apiURL, err := url.JoinPath(p.cfg.TrayaAPIServerURL.String(), "community-user")
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to fetch user data from Traya API"))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to read Traya API response body"))
	}

	p.logger.Debug("received raw response from Traya API",
		slog.Int("status", resp.StatusCode),
		slog.String("body", string(body)))

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fault.Wrap(ErrInvalidToken, fctx.With(ctx))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fault.Wrap(ErrTrayaAPIFailed,
			fctx.With(ctx),
			fmsg.WithDesc("unexpected status code", fmt.Sprintf("Traya API returned status %d", resp.StatusCode)))
	}

	var userData TrayaUserResponse
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx), fmsg.With("failed to decode Traya API response"))
	}

	return &userData, nil
}

func (p *Provider) getOrCreateAccount(
	ctx context.Context,
	trayaUserID string,
	email mail.Address,
	firstName string,
	lastName *string,
	phoneNumber string,
) (*account.Account, error) {
	authMethod, authMethodExists, err := p.authRepo.LookupByIdentifier(ctx, service, trayaUserID)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if authMethodExists {
		return &authMethod.Account, nil
	}

	existingAccount, emailExists, err := p.emailRepo.LookupAccount(ctx, email)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if emailExists {
		_, err = p.authRepo.Create(ctx, existingAccount.ID, service, authentication.TokenTypeTraya, trayaUserID, trayaUserID, nil, authentication.WithName(email.Address))
		if err != nil {
			return nil, fault.Wrap(err, fctx.With(ctx))
		}
		return &existingAccount.Account, nil
	}

	name := firstName
	if lastName != nil && *lastName != "" {
		name = fmt.Sprintf("%s %s", firstName, *lastName)
	}

	// Use temporary handle that user will replace during onboarding
	// Generate short unique handle: temp_ (5 chars) + xid (20 chars) = 25 chars (under 30 limit)
	handle := fmt.Sprintf("temp_%s", xid.New().String())

	newAccount, err := p.register.GetOrCreateViaEmail(ctx, service, email.Address, trayaUserID, trayaUserID, handle, name, email, true)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return newAccount, nil
}

func generateHandle(firstName string, phoneNumber string) string {
	namePrefix := firstName
	if len(namePrefix) > 4 {
		namePrefix = namePrefix[:4]
	}
	namePrefix = strings.ToLower(namePrefix)

	phoneDigits := strings.TrimPrefix(phoneNumber, "+")
	phoneDigits = strings.ReplaceAll(phoneDigits, " ", "")
	phoneDigits = strings.ReplaceAll(phoneDigits, "-", "")

	phonePrefix := ""
	if len(phoneDigits) >= 2 {
		phonePrefix = phoneDigits[:2]
	}

	randomDigits := fmt.Sprintf("%04d", rand.Intn(10000))

	return fmt.Sprintf("%s%s%s", namePrefix, phonePrefix, randomDigits)
}

func (p *Provider) ensureChannelMemberships(ctx context.Context, accountID account.AccountID, gender string, orderCount int, latestOrderDate string) error {
	normalizedGender := normalizeGender(gender)

	isWithin60Days, err := isLastOrderWithin60Days(latestOrderDate)
	if err != nil {
		p.logger.Warn("failed to parse latest order date",
			slog.String("date", latestOrderDate),
			slog.String("error", err.Error()))
		isWithin60Days = false
	}

	// Step 1: Build target channels set based on current gender/orderCount
	targetChannels := make(map[string]bool)

	if isWithin60Days {
		for _, rule := range cohortChannelRules {
			if rule.gender == normalizedGender && rule.orderCount == orderCount {
				targetChannels[rule.channelSlug] = true
			}
		}

		if orderCount > 8 {
			if normalizedGender == "male" {
				targetChannels["month-8-plus-legends"] = true
			} else if normalizedGender == "female" {
				targetChannels["month-8-plus-queens"] = true
			}
		}
	}

	if topicChannels, ok := topicChannelsByGender[normalizedGender]; ok {
		for _, channelSlug := range topicChannels {
			targetChannels[channelSlug] = true
		}
	}

	// Step 2: Get all managed channel slugs
	managedChannelSlugs := getAllManagedChannelSlugs()

	// Step 3: Get current memberships for the user
	currentMemberships, err := p.membershipRepo.ListByAccount(ctx, accountID)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	// Step 4: Build map of current managed channels (channelID -> slug)
	currentManagedChannels := make(map[xid.ID]string)
	for _, membership := range currentMemberships {
		ch, err := p.channelRepo.Get(ctx, channel.ChannelID(membership.ChannelID))
		if err != nil {
			p.logger.Warn("failed to get channel for membership",
				slog.String("channel_id", membership.ChannelID.String()),
				slog.String("error", err.Error()))
			continue
		}

		if managedChannelSlugs[ch.Slug] {
			currentManagedChannels[membership.ChannelID] = ch.Slug
		}
	}

	// Step 5: Remove memberships that should no longer exist
	for channelID, slug := range currentManagedChannels {
		if !targetChannels[slug] {
			err := p.membershipRepo.Remove(ctx, channelID, accountID)
			if err != nil {
				p.logger.Warn("failed to remove channel membership",
					slog.String("channel", slug),
					slog.String("account_id", accountID.String()),
					slog.String("error", err.Error()))
				continue
			}

			p.logger.Info("removed user from channel",
				slog.String("channel", slug),
				slog.String("account_id", accountID.String()),
				slog.Int("order_count", orderCount),
				slog.String("gender", gender))
		}
	}

	// Step 6: Add memberships that should exist but don't
	for channelSlug := range targetChannels {
		alreadyMember := false
		for _, currentSlug := range currentManagedChannels {
			if currentSlug == channelSlug {
				alreadyMember = true
				break
			}
		}

		if !alreadyMember {
			ch, err := p.channelRepo.GetBySlug(ctx, channelSlug)
			if err != nil {
				p.logger.Warn("channel not found",
					slog.String("slug", channelSlug),
					slog.String("error", err.Error()))
				continue
			}

			_, err = p.membershipRepo.Add(ctx, xid.ID(ch.ID), accountID, channel_membership.RoleMember)
			if err != nil {
				p.logger.Warn("failed to add channel membership",
					slog.String("channel", channelSlug),
					slog.String("account_id", accountID.String()),
					slog.String("error", err.Error()))
				continue
			}

			p.logger.Info("added user to channel",
				slog.String("channel", channelSlug),
				slog.String("account_id", accountID.String()),
				slog.Int("order_count", orderCount),
				slog.String("gender", gender))
		}
	}

	return nil
}

func normalizeGender(gender string) string {
	gender = strings.ToUpper(strings.TrimSpace(gender))
	switch gender {
	case "M":
		return "male"
	case "F":
		return "female"
	default:
		return strings.ToLower(gender)
	}
}

func isLastOrderWithin60Days(latestOrderDate string) (bool, error) {
	if latestOrderDate == "" {
		return false, nil
	}

	orderDate, err := time.Parse(time.RFC3339, latestOrderDate)
	if err != nil {
		orderDate, err = time.Parse("2006-01-02", latestOrderDate)
		if err != nil {
			return false, err
		}
	}

	daysSinceOrder := time.Since(orderDate).Hours() / 24
	return daysSinceOrder <= 60, nil
}

func (p *Provider) Enabled(ctx context.Context) (bool, error) {
	return p.cfg.TrayaAPIServerURL.String() != "", nil
}
