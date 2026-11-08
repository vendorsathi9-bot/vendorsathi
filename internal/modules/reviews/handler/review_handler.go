package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/reviews/model"
	"vendorsaathi/internal/modules/reviews/service"
	"vendorsaathi/internal/platform/httpx"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: svc}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	if id, ok := val.(uuid.UUID); ok {
		return id, true
	}
	if str, ok := val.(string); ok {
		id, err := uuid.Parse(str)
		if err == nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

// CreateReview POST /reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	var input model.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid review payload", nil)
		return
	}

	output, err := h.service.CreateReview(c.Request.Context(), customerID, input)
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondCreated(c, output)
}

// GetVendorReviews GET /reviews/vendor/:vendorId
func (h *ReviewHandler) GetVendorReviews(c *gin.Context) {
	vendorIDStr := c.Param("vendorId")
	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid vendor id", nil)
		return
	}

	output, err := h.service.GetVendorReviews(c.Request.Context(), vendorID)
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondOK(c, output)
}

// GetAllReviews GET /reviews/admin
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	output, err := h.service.GetAllReviews(c.Request.Context())
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}
	httpx.RespondOK(c, output)
}

// ModerateReview PATCH /reviews/:id/moderate
func (h *ReviewHandler) ModerateReview(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid review id", nil)
		return
	}

	var input model.ModerateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid input", nil)
		return
	}

	if err := h.service.ModerateReview(c.Request.Context(), reviewID, input.IsHidden); err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondOK(c, gin.H{"status": "updated", "is_hidden": input.IsHidden})
}

// GetMyLoyalty GET /loyalty/me
func (h *ReviewHandler) GetMyLoyalty(c *gin.Context) {
	customerID, ok := getUserID(c)
	if !ok {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	output, err := h.service.GetLoyaltySummary(c.Request.Context(), customerID)
	if err != nil {
		httpx.ErrorInternal(c, err.Error())
		return
	}

	httpx.RespondOK(c, output)
}
