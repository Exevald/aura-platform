package adapter

import (
	"context"
	"fmt"
	port "publication/internal/publication/app/adapter"
	"uuid"

	api "shortlink/api/server/internalapi"
)

const managedLinkOwnerID = "84c7d60e-6f3e-4f09-91c2-82774c12fc88"

func NewShortlinkService(client api.ShortlinkServiceClient) port.ShortlinkService {
	return shortlinkAdapter{client}
}

type shortlinkAdapter struct {
	client api.ShortlinkServiceClient
}

func (s shortlinkAdapter) Ensure(
	ctx context.Context,
	id uuid.UUID,
	target string,
) (linkID uuid.UUID, shortPath string, err error) {
	r, err := s.client.EnsureLink(
		ctx,
		&api.EnsureLinkRequest{
			OwnerUserId:    managedLinkOwnerID,
			IdempotencyKey: id.String(),
			TargetUrl:      target,
		})
	if err != nil {
		return uuid.Nil(), "", err
	}
	linkID, err = uuid.Parse(r.LinkId)
	if err != nil {
		return uuid.Nil(), "", fmt.Errorf("invalid shortlink ID %q: %w", r.LinkId, err)
	}
	return linkID, r.ShortPath, nil
}

func (s shortlinkAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.client.DeleteOwnedLink(
		ctx,
		&api.DeleteOwnedLinkRequest{
			OwnerUserId:    managedLinkOwnerID,
			IdempotencyKey: id.String(),
		})
	return err
}
