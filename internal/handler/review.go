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

func (h *Handler) CreateReview(ctx context.Context, req *api.CreateReviewReq, params api.CreateReviewParams) (api.CreateReviewRes, error) {
	reviewer := getUserID(ctx)

	repoType, err := toRepositoryReviewType(req.Type)
	if err != nil {
		return &api.CreateReviewBadRequest{}, nil
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
			return &api.CreateReviewNotFound{}, nil
		case repository.ErrReviewerNotFound:
			return &api.CreateReviewNotFound{}, nil
		case repository.ErrReviewAlreadyExists:
			return &api.CreateReviewConflict{}, nil
		case repository.ErrInvalidReviewType, repository.ErrInvalidReviewWeight:
			return &api.CreateReviewBadRequest{}, nil
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

// DeleteReview implements DELETE /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) DeleteReview(ctx context.Context, params api.DeleteReviewParams) error {
	reviewer := getUserID(ctx)

	if err := h.repo.DeleteReview(ctx, params.TicketId, params.NoteId, params.ReviewId, reviewer); err != nil {
		switch err {
		case repository.ErrReviewNotFound:
			return echo.NewHTTPError(http.StatusNotFound, "review not found")
		case repository.ErrReviewForbidden:
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		default:
			return fmt.Errorf("delete review in repository: %w", err)
		}
	}

	return nil
}

// UpdateReview implements PUT /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) UpdateReview(ctx context.Context, req api.OptUpdateReviewReq, params api.UpdateReviewParams) (api.UpdateReviewRes, error) {
	reviewer := getUserID(ctx)

	repoParams := repository.UpdateReviewParams{
		Type:       "",
		TypeSet:    false,
		Weight:     0,
		WeightSet:  false,
		Comment:    sql.NullString{String: "", Valid: false},
		CommentSet: false,
	}
	if req.Set {
		if req.Value.Type.Set {
			repoType, err := toRepositoryReviewType(req.Value.Type.Value)
			if err != nil {
				return nil, err
			}
			repoParams.Type = repoType
			repoParams.TypeSet = true
		}

		if req.Value.Weight.Set {
			repoParams.Weight = req.Value.Weight.Value
			repoParams.WeightSet = true
		}

		if req.Value.Comment.Set {
			repoParams.Comment = sql.NullString{String: req.Value.Comment.Value, Valid: true}
			repoParams.CommentSet = true
		}
	}

	_, err := h.repo.UpdateReview(ctx, params.TicketId, params.NoteId, params.ReviewId, reviewer, repoParams)
	if err != nil {
		switch err {
		case repository.ErrReviewNotFound:
			return &api.UpdateReviewNotFound{}, nil
		case repository.ErrReviewForbidden:
			return &api.UpdateReviewForbidden{}, nil
		case repository.ErrReviewerNotFound:
			return &api.UpdateReviewNotFound{}, nil
		case repository.ErrInvalidReviewType, repository.ErrInvalidReviewWeight:
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return nil, fmt.Errorf("update review in repository: %w", err)
		}
	}

	return &api.UpdateReviewOK{}, nil
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
