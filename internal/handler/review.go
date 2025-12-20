package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

func (h *Handler) TicketsTicketIdNotesNoteIdReviewsPost(ctx context.Context, req *api.TicketsTicketIdNotesNoteIdReviewsPostReq, params api.TicketsTicketIdNotesNoteIdReviewsPostParams) (*api.Review, error) {
	reviewer, ok := traqIDFromContext(ctx)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	repoType, err := toRepositoryReviewType(req.Type)
	if err != nil {
		return nil, err
	}

	comment := sql.NullString{String: "", Valid: false}
	if req.Comment.Set {
		comment = sql.NullString{String: req.Comment.Value, Valid: true}
	}

	repoReview, err := h.repo.CreateReview(ctx, params.TicketId, params.NoteId, reviewer, repository.CreateReviewParams{
		Type:      repoType,
		Weight:    req.Weight.Value,
		WeightSet: req.Weight.Set,
		Comment:   comment,
	})
	if err != nil {
		switch err {
		case repository.ErrNoteNotFound:
			return nil, echo.NewHTTPError(http.StatusNotFound, "note not found")
		case repository.ErrReviewerNotFound:
			return nil, echo.NewHTTPError(http.StatusNotFound, "reviewer not found")
		case repository.ErrReviewAlreadyExists:
			return nil, echo.NewHTTPError(http.StatusConflict, "already reviewed")
		case repository.ErrInvalidReviewType, repository.ErrInvalidReviewWeight:
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return nil, fmt.Errorf("create review in repository: %w", err)
		}
	}

	apiReview, err := convertRepositoryReview(repoReview)
	if err != nil {
		return nil, fmt.Errorf("convert review: %w", err)
	}

	return apiReview, nil
}
func toRepositoryReviewType(t api.ReviewType) (string, error) {
	switch t {
	case api.ReviewTypeApprove:
		return "approve", nil
	case api.ReviewTypeChangeRequest:
		return "cr", nil
	case api.ReviewTypeComment:
		return "comment", nil
	case api.ReviewTypeSystem:
		return "system", nil
	default:
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid review type")
	}
}

func convertRepositoryReview(review *repository.Review) (*api.Review, error) {
	reviewType, err := toAPIReviewType(review.Type)
	if err != nil {
		return nil, err
	}

	reviewStatus, err := toAPIReviewStatus(review.Status)
	if err != nil {
		return nil, err
	}

	return &api.Review{
		ID:       review.ID,
		NoteID:   review.NoteID,
		Reviewer: review.Author,
		Type:     reviewType,
		Weight:   review.Weight,
		Status:   reviewStatus,
		Comment: api.OptString{
			Value: review.Comment.String,
			Set:   review.Comment.Valid,
		},
		CreatedAt: api.OptDateTime{Value: review.CreatedAt, Set: true},
		UpdatedAt: api.OptDateTime{Value: review.UpdatedAt, Set: true},
	}, nil
}

func toAPIReviewType(t string) (api.ReviewType, error) {
	switch t {
	case "approve":
		return api.ReviewTypeApprove, nil
	case "cr":
		return api.ReviewTypeChangeRequest, nil
	case "comment":
		return api.ReviewTypeComment, nil
	case "system":
		return api.ReviewTypeSystem, nil
	default:
		return "", fmt.Errorf("unknown review type: %s", t)
	}
}

func toAPIReviewStatus(status string) (api.ReviewStatus, error) {
	switch status {
	case "active":
		return api.ReviewStatusActive, nil
	case "stale":
		return api.ReviewStatusStale, nil
	default:
		return "", fmt.Errorf("unknown review status: %s", status)
	}
}
