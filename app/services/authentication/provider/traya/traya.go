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

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/account_querier"
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

// channelsByKitCount maps kit count thresholds to the list of channels users should be added to
// Users are added to all channels where their kit count meets or exceeds the threshold
var channelsByKitCount = []struct {
	minKitCount int
	channels    []string
}{
	{0, []string{"general"}},
	{1, []string{"general", "month-0-to-3"}},
	{4, []string{"general", "month-3-to-6"}},
	{7, []string{"general", "month-6-to-9"}},
	{10, []string{"general", "month-9-to-12"}},
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

	kitCount := userData.Case.LatestOrderCount
	if kitCount == 0 {
		kitCount = userData.TotalKitCount
	}

	acc, err := p.getOrCreateAccount(ctx, userData.User.ID, *emailAddress, userData.User.FirstName, userData.User.LastName, userData.User.PhoneNumber)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	if err := p.ensureChannelMemberships(ctx, acc.ID, kitCount); err != nil {
		p.logger.Warn("failed to ensure channel memberships",
			slog.String("account_id", acc.ID.String()),
			slog.Int("kit_count", kitCount),
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

	handle := generateHandle(firstName, phoneNumber)

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

func (p *Provider) ensureChannelMemberships(ctx context.Context, accountID account.AccountID, kitCount int) error {
	// Collect all channels the user should be added to based on their kit count
	channelsToAdd := make(map[string]bool)

	for _, entry := range channelsByKitCount {
		if kitCount >= entry.minKitCount {
			for _, channelSlug := range entry.channels {
				channelsToAdd[channelSlug] = true
			}
		}
	}

	// Add user to all eligible channels
	for channelSlug := range channelsToAdd {
		channel, err := p.channelRepo.GetBySlug(ctx, channelSlug)
		if err != nil {
			p.logger.Warn("channel not found",
				slog.String("slug", channelSlug),
				slog.String("error", err.Error()))
			continue
		}

		exists, err := p.membershipRepo.CheckMembership(ctx, xid.ID(channel.ID), accountID)
		if err != nil {
			return fault.Wrap(err, fctx.With(ctx))
		}

		if !exists {
			_, err = p.membershipRepo.Add(ctx, xid.ID(channel.ID), accountID, channel_membership.RoleMember)
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
				slog.Int("kit_count", kitCount))
		}
	}

	return nil
}

func (p *Provider) Enabled(ctx context.Context) (bool, error) {
	return p.cfg.TrayaAPIServerURL.String() != "", nil
}
