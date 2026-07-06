package convert_job

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"

	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/asset/asset_querier"
	"github.com/Southclaws/storyden/app/resources/asset/asset_writer"
	"github.com/Southclaws/storyden/app/services/asset/asset_upload"
	"github.com/Southclaws/storyden/internal/infrastructure/object"
	"github.com/Southclaws/storyden/internal/mime"
)

type Consumer struct {
	assetQuerier *asset_querier.Querier
	assetWriter  *asset_writer.Writer
	objects      object.Storer
}

func New(
	assetQuerier *asset_querier.Querier,
	assetWriter *asset_writer.Writer,
	objects object.Storer,
) *Consumer {
	return &Consumer{
		assetQuerier: assetQuerier,
		assetWriter:  assetWriter,
		objects:      objects,
	}
}

func (c *Consumer) Convert(ctx context.Context, id asset.AssetID) error {
	ast, err := c.assetQuerier.GetByID(ctx, id)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	if !asset_upload.NeedsJPEGConversion(ast.MIME) {
		return nil
	}

	path := asset.BuildAssetPath(ast.Name)

	original, _, err := c.objects.Read(ctx, path)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	converted, size, err := asset_upload.ConvertToJPEG(original)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	if err := c.objects.Write(ctx, path, converted, size); err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	if _, err := c.assetWriter.UpdateContent(ctx, id, int(size), mime.New("image/jpeg")); err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	return nil
}
