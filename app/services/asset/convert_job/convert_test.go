package convert_job_test

import (
	"bytes"
	"context"
	"image"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account/account_writer"
	"github.com/Southclaws/storyden/app/resources/account/role"
	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/asset/asset_querier"
	"github.com/Southclaws/storyden/app/resources/seed"
	"github.com/Southclaws/storyden/app/services/asset/asset_upload"
	"github.com/Southclaws/storyden/app/services/asset/convert_job"
	"github.com/Southclaws/storyden/app/services/authentication/session"
	"github.com/Southclaws/storyden/internal/infrastructure/object"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestConvertJob_HEICToJPEG(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		aw *account_writer.Writer,
		uploader *asset_upload.Uploader,
		querier *asset_querier.Querier,
		objects object.Storer,
		consumer *convert_job.Consumer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			ctx, acc := e2e.WithAccount(root, aw, seed.Account_001_Odin)
			ctx = session.WithAccount(ctx, *acc, role.Roles{})

			raw, err := os.ReadFile("../asset_upload/testdata/sample.heic")
			r.NoError(err)

			ast, err := uploader.Upload(ctx, bytes.NewReader(raw), int64(len(raw)), asset.NewFilename("photo"), asset_upload.Options{})
			r.NoError(err)
			a.Equal("image/heic", ast.MIME.String())

			// The original is stored as-is; conversion happens off the request path.
			original, _, err := objects.Read(ctx, asset.BuildAssetPath(ast.Name))
			r.NoError(err)
			originalBytes, err := io.ReadAll(original)
			r.NoError(err)
			a.Equal(raw, originalBytes)

			r.NoError(consumer.Convert(ctx, ast.ID))

			updated, err := querier.GetByID(ctx, ast.ID)
			r.NoError(err)
			a.Equal("image/jpeg", updated.MIME.String())

			stored, size, err := objects.Read(ctx, asset.BuildAssetPath(updated.Name))
			r.NoError(err)
			storedBytes, err := io.ReadAll(stored)
			r.NoError(err)

			_, format, err := image.Decode(bytes.NewReader(storedBytes))
			r.NoError(err)
			a.Equal("jpeg", format)
			a.Equal(int64(len(storedBytes)), size)
			a.Equal(len(storedBytes), updated.Size)
		}))
	}))
}

func TestConvertJob_SkipsAlreadyConverted(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		aw *account_writer.Writer,
		uploader *asset_upload.Uploader,
		querier *asset_querier.Querier,
		consumer *convert_job.Consumer,
	) {
		lc.Append(fx.StartHook(func() {
			r := require.New(t)
			a := assert.New(t)

			ctx, acc := e2e.WithAccount(root, aw, seed.Account_002_Frigg)
			ctx = session.WithAccount(ctx, *acc, role.Roles{})

			raw, err := os.ReadFile("../asset_upload/testdata/sample.heic")
			r.NoError(err)

			ast, err := uploader.Upload(ctx, bytes.NewReader(raw), int64(len(raw)), asset.NewFilename("photo2"), asset_upload.Options{})
			r.NoError(err)

			r.NoError(consumer.Convert(ctx, ast.ID))
			// Redelivery is a safe no-op once MIME is already image/jpeg.
			r.NoError(consumer.Convert(ctx, ast.ID))

			updated, err := querier.GetByID(ctx, ast.ID)
			r.NoError(err)
			a.Equal("image/jpeg", updated.MIME.String())
		}))
	}))
}
