package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vendorsaathi/internal/modules/products/model"
	"vendorsaathi/internal/modules/products/service"
	"vendorsaathi/internal/platform/httpx"
	"vendorsaathi/internal/shared"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var input model.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.productService.CreateCategory(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *ProductHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid category id", nil)
		return
	}

	output, err := h.productService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) GetCategories(c *gin.Context) {
	activeOnly := true
	if activeStr := c.Query("active_only"); activeStr != "" {
		if parsed, err := strconv.ParseBool(activeStr); err == nil {
			activeOnly = parsed
		}
	}

	output, err := h.productService.GetCategories(c.Request.Context(), activeOnly)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid category id", nil)
		return
	}

	var input model.UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.productService.UpdateCategory(c.Request.Context(), id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	var input model.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	vendorID := userID

	output, err := h.productService.CreateProduct(c.Request.Context(), vendorID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondCreated(c, output)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	output, err := h.productService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) GetMyProducts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	params := httpx.PaginationParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		httpx.ErrorBadRequest(c, "invalid query parameters", nil)
		return
	}

	activeOnly := true
	if activeStr := c.Query("active_only"); activeStr != "" {
		if parsed, err := strconv.ParseBool(activeStr); err == nil {
			activeOnly = parsed
		}
	}

	output, err := h.productService.GetMyProducts(c.Request.Context(), vendorID, params.Limit, params.Cursor, activeOnly)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondPaginated(c, output.Data, output.NextCursor, output.HasMore)
}

func (h *ProductHandler) GetMyActiveProducts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	output, err := h.productService.GetMyActiveProducts(c.Request.Context(), vendorID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	var input model.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ErrorBadRequest(c, "invalid request body", nil)
		return
	}

	output, err := h.productService.UpdateProduct(c.Request.Context(), vendorID, productID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) ArchiveProduct(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		httpx.ErrorUnauthorized(c, "user not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		httpx.ErrorInternal(c, "invalid user id")
		return
	}

	rolesVal, _ := c.Get("roles")
	userRoles := rolesVal.([]string)

	isVendor := false
	for _, r := range userRoles {
		if r == "VENDOR" {
			isVendor = true
			break
		}
	}
	if !isVendor {
		httpx.ErrorForbidden(c, "vendor role required")
		return
	}

	vendorID := userID

	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid product id", nil)
		return
	}

	output, err := h.productService.ArchiveProduct(c.Request.Context(), vendorID, productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid category id", nil)
		return
	}

	activeOnly := true
	if activeStr := c.Query("active_only"); activeStr != "" {
		if parsed, err := strconv.ParseBool(activeStr); err == nil {
			activeOnly = parsed
		}
	}

	output, err := h.productService.GetProductsByCategory(c.Request.Context(), categoryID, activeOnly)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) GetVendorProducts(c *gin.Context) {
	vendorIDStr := c.Param("vendorId")
	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		httpx.ErrorBadRequest(c, "invalid vendor id", nil)
		return
	}

	output, err := h.productService.GetVendorProducts(c.Request.Context(), vendorID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.RespondOK(c, output)
}

func (h *ProductHandler) handleError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "NOT_FOUND":
			httpx.ErrorNotFound(c, appErr.Message)
		case "FORBIDDEN":
			httpx.ErrorForbidden(c, appErr.Message)
		case "CONFLICT":
			httpx.ErrorConflict(c, appErr.Message, nil)
		case "VALIDATION_ERROR":
			var details []httpx.ErrorDetail
			for field, reason := range appErr.Details {
				details = append(details, httpx.ErrorDetail{Field: field, Reason: reason})
			}
			httpx.ErrorUnprocessable(c, appErr.Message, details)
		default:
			httpx.ErrorInternal(c, appErr.Message)
		}
		return
	}

	if errors.Is(err, shared.ErrNotFound) {
		httpx.ErrorNotFound(c, "resource not found")
		return
	}
	if errors.Is(err, shared.ErrForbidden) {
		httpx.ErrorForbidden(c, "access forbidden")
		return
	}
	if errors.Is(err, shared.ErrConflict) {
		httpx.ErrorConflict(c, err.Error(), nil)
		return
	}
	if errors.Is(err, shared.ErrValidationFailed) {
		httpx.ErrorUnprocessable(c, "validation failed", nil)
		return
	}

	httpx.ErrorInternal(c, "internal server error")
}
