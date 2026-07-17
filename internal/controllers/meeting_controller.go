package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/meeting-minutes-ai/internal/dto"
	"github.com/yourusername/meeting-minutes-ai/internal/services"
)

// MeetingController handles meeting endpoints
type MeetingController struct {
	meetingService *services.MeetingService
}

// NewMeetingController creates a new meeting controller
func NewMeetingController(meetingService *services.MeetingService) *MeetingController {
	return &MeetingController{meetingService: meetingService}
}

// CreateMeeting handles creating a new meeting
// @Summary Create a new meeting
// @Tags Meetings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateMeetingRequest true "Meeting data"
// @Success 201 {object} dto.APIResponse
// @Router /api/meetings [post]
func (c *MeetingController) CreateMeeting(ctx *gin.Context) {
	var req dto.CreateMeetingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	organizerID, _ := ctx.Get("user_id")
	meeting, err := c.meetingService.CreateMeeting(&req, organizerID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Rapat berhasil dibuat",
		Data:    meeting,
	})
}

// GetMeeting handles getting meeting details
// @Summary Get meeting details
// @Tags Meetings
// @Produce json
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/{id} [get]
func (c *MeetingController) GetMeeting(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   "Invalid meeting ID",
		})
		return
	}

	detail, err := c.meetingService.GetMeetingDetail(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    detail,
	})
}

// ListMeetings handles listing meetings
// @Summary List all meetings
// @Tags Meetings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings [get]
func (c *MeetingController) ListMeetings(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	meetings, total, err := c.meetingService.ListMeetings(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: gin.H{
			"meetings": meetings,
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

// UploadAudio handles audio upload for a meeting
// @Summary Upload meeting audio
// @Tags Meetings
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param meeting_id formData int true "Meeting ID"
// @Param audio formData file true "Audio file"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/upload-audio [post]
func (c *MeetingController) UploadAudio(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.PostForm("meeting_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   "Invalid meeting ID",
		})
		return
	}

	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   "Audio file is required",
		})
		return
	}

	if err := c.meetingService.UploadAudio(uint(meetingID), file); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Audio berhasil diupload, sedang diproses",
	})
}

// ProcessTranscript handles processing transcript
// @Summary Process meeting transcript
// @Tags Meetings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ProcessTranscriptRequest true "Transcript data"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/process-transcript [post]
func (c *MeetingController) ProcessTranscript(ctx *gin.Context) {
	var req dto.ProcessTranscriptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	meeting, err := c.meetingService.ProcessTranscript(req.MeetingID, req.Transcript)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Notulensi berhasil diproses",
		Data:    meeting,
	})
}

// ExportMeeting handles exporting meeting minutes
// @Summary Export meeting minutes
// @Tags Export
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ExportRequest true "Export data"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/export [post]
func (c *MeetingController) ExportMeeting(ctx *gin.Context) {
	var req dto.ExportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	path, err := c.meetingService.ExportMeeting(req.MeetingID, req.Format)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Serve file download
	contentType := "application/octet-stream"
	switch req.Format {
	case "pdf":
		contentType = "application/pdf"
	case "word", "docx":
		contentType = "application/msword"
	}

	ctx.Header("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	ctx.Header("Content-Type", contentType)
	ctx.File(path)
}

// SendEmail handles sending meeting minutes via email
// @Summary Send meeting minutes via email
// @Tags Export
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SendEmailRequest true "Email data"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/send-email [post]
func (c *MeetingController) SendEmail(ctx *gin.Context) {
	var req dto.SendEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	format := req.Format
	if format == "" {
		format = "pdf"
	}

	if err := c.meetingService.SendMeetingEmail(req.MeetingID, req.Recipients, format); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Email berhasil dikirim",
	})
}

// UpdateMeeting handles updating a meeting's basic info
// @Summary Update meeting info (title, date, location, status)
// @Tags Meetings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Param request body dto.UpdateMeetingRequest true "Meeting update data"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/{id} [put]
func (c *MeetingController) UpdateMeeting(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   "Invalid meeting ID",
		})
		return
	}

	var req dto.UpdateMeetingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	meeting, err := c.meetingService.UpdateMeeting(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Rapat berhasil diperbarui",
		Data:    meeting,
	})
}

// DeleteMeeting handles deleting a meeting
// @Summary Delete a meeting
// @Tags Meetings
// @Produce json
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} dto.APIResponse
// @Router /api/meetings/{id} [delete]
func (c *MeetingController) DeleteMeeting(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   "Invalid meeting ID",
		})
		return
	}

	if err := c.meetingService.DeleteMeeting(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Rapat berhasil dihapus",
	})
}

// GetDashboard handles dashboard statistics
// @Summary Get dashboard statistics
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse
// @Router /api/dashboard [get]
func (c *MeetingController) GetDashboard(ctx *gin.Context) {
	stats, err := c.meetingService.GetDashboardStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    stats,
	})
}