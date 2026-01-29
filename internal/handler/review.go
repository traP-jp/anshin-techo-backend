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

	role, err := h.repo.GetUserRoleByTraqID(ctx, reviewer)
	if err != nil {
		return nil, fmt.Errorf("get user role: %w", err)
	}

	repoType, err := toRepositoryReviewType(req.Type)
	if err != nil {
		return &api.CreateReviewBadRequest{}, nil
	}

	comment := sql.NullString{String: req.Comment, Valid: true}

	repoReview, err := h.repo.CreateReview(ctx, params.TicketId, params.NoteId, reviewer, repository.CreateReviewParams{
		Type:    repoType,
		Weight:  req.Weight,
		Comment: comment,
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

	apiReview, err := convertRepositoryReview(repoReview, role)
	if err != nil {
		return nil, fmt.Errorf("convert review: %w", err)
	}

	return apiReview, nil
}

// DeleteReview implements DELETE /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) DeleteReview(ctx context.Context, params api.DeleteReviewParams) (api.DeleteReviewRes, error) {
	reviewer := getUserID(ctx)

	if err := h.repo.DeleteReview(ctx, params.TicketId, params.NoteId, params.ReviewId, reviewer); err != nil {
		switch err {
		case repository.ErrReviewNotFound:
			return &api.DeleteReviewNotFound{}, nil
		case repository.ErrReviewForbidden:
			return &api.DeleteReviewForbidden{}, nil
		default:
			return nil, fmt.Errorf("delete review in repository: %w", err)
		}
	}

	return &api.DeleteReviewNoContent{}, nil
}

// UpdateReview implements PUT /tickets/{ticketId}/notes/{noteId}/reviews/{reviewId} operation.
func (h *Handler) UpdateReview(ctx context.Context, req api.OptUpdateReviewReq, params api.UpdateReviewParams) (api.UpdateReviewRes, error) {
	reviewer := getUserID(ctx)

	repoType, err := toRepositoryReviewType(req.Value.Type)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	repoParams := repository.UpdateReviewParams{
		TypeSet:    true,
		WeightSet:  true,
		CommentSet: true,
		Type:       repoType,
		Weight:     req.Value.Weight,
		Comment:    sql.NullString{String: req.Value.Comment, Valid: true},
	}

	_, err = h.repo.UpdateReview(ctx, params.TicketId, params.NoteId, params.ReviewId, reviewer, repoParams)
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

func convertRepositoryReview(review *repository.Review, role string) (*api.Review, error) {
	reviewType, err := toAPIReviewType(review.Type)
	if err != nil {
		return nil, err
	}

	reviewStatus, err := toAPIReviewStatus(review.Status)
	if err != nil {
		return nil, err
	}

	safeComment := ApplyCensorIfNeed(role, review.Comment.String)

	return &api.Review{
		ID:        review.ID,
		NoteID:    review.NoteID,
		Reviewer:  review.Author,
		Type:      reviewType,
		Weight:    review.Weight,
		Status:    reviewStatus,
		Comment:   safeComment,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
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
