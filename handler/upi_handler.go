package handler

import (
	"errors"
	"strconv"

	"github.com/banking-superapp/upi-service/model"
	"github.com/banking-superapp/upi-service/service"
	"github.com/gofiber/fiber/v2"
)

type UPIHandler struct {
	svc service.UPIService
}

func NewUPIHandler(svc service.UPIService) *UPIHandler {
	return &UPIHandler{svc: svc}
}

func (h *UPIHandler) CreateVPA(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	var req model.CreateVPARequest
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	vpa, err := h.svc.CreateVPA(c.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrVPAExists) {
			return respond(c, fiber.StatusConflict, nil, err.Error())
		}
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusCreated, vpa, "")
}

func (h *UPIHandler) GetVPAs(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	vpas, err := h.svc.GetVPAs(c.Context(), userID)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, vpas, "")
}

func (h *UPIHandler) ValidateVPA(c *fiber.Ctx) error {
	var req model.ValidateVPARequest
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	result, err := h.svc.ValidateVPA(c.Context(), req.VPA)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, result, "")
}

func (h *UPIHandler) Pay(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	var req model.UPIPayRequest
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	txn, err := h.svc.Pay(c.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			return respond(c, fiber.StatusBadRequest, nil, err.Error())
		}
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusCreated, txn, "")
}

func (h *UPIHandler) Collect(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	var req model.CollectRequestInput
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	cr, err := h.svc.Collect(c.Context(), userID, &req)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusCreated, cr, "")
}

func (h *UPIHandler) GetTransactions(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
	txns, total, err := h.svc.GetTransactions(c.Context(), userID, page, limit)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, fiber.Map{"transactions": txns, "total": total, "page": page}, "")
}

func (h *UPIHandler) CreateMandate(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	var req model.CreateMandateRequest
	if err := c.BodyParser(&req); err != nil {
		return respond(c, fiber.StatusBadRequest, nil, "invalid request body")
	}
	mandate, err := h.svc.CreateMandate(c.Context(), userID, &req)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusCreated, mandate, "")
}

func (h *UPIHandler) GetMandates(c *fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	mandates, err := h.svc.GetMandates(c.Context(), userID)
	if err != nil {
		return respond(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return respond(c, fiber.StatusOK, mandates, "")
}

func respond(c *fiber.Ctx, status int, data interface{}, errMsg string) error {
	if errMsg != "" {
		return c.Status(status).JSON(fiber.Map{"success": false, "error": errMsg})
	}
	return c.Status(status).JSON(fiber.Map{"success": true, "data": data})
}
