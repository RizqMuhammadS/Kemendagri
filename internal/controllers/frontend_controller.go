package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// FrontendController handles web page rendering
type FrontendController struct{}

// NewFrontendController creates a new frontend controller
func NewFrontendController() *FrontendController {
	return &FrontendController{}
}

// LoginPage renders the login page
func (fc *FrontendController) LoginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Sistem Notulensi",
	})
}

// RegisterPage renders the register page
func (fc *FrontendController) RegisterPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "register.html", gin.H{
		"title": "Register - Sistem Notulensi",
	})
}

// DashboardPage renders the dashboard page
func (fc *FrontendController) DashboardPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard - Sistem Notulensi",
	})
}

// MeetingsPage renders the meetings list page
func (fc *FrontendController) MeetingsPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "meetings.html", gin.H{
		"title": "Daftar Rapat - Sistem Notulensi",
	})
}

// CreateMeetingPage renders the create meeting page
func (fc *FrontendController) CreateMeetingPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "create_meeting.html", gin.H{
		"title": "Buat Rapat Baru - Sistem Notulensi",
	})
}

// MeetingDetailPage renders the meeting detail page
func (fc *FrontendController) MeetingDetailPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "meeting_detail.html", gin.H{
		"title": "Detail Rapat - Sistem Notulensi",
	})
}

// ServeFrontend serves the main SPA layout with content determined by the URL path
func (fc *FrontendController) ServeFrontend(ctx *gin.Context) {
	path := ctx.Request.URL.Path

	// Map paths to template names
	templateMap := map[string]string{
		"/":               "dashboard.html",
		"/dashboard":      "dashboard.html",
		"/meetings":       "meetings.html",
		"/meetings/create": "create_meeting.html",
	}

	templateName := templateMap[path]
	if templateName == "" {
		if len(path) > 10 && path[:10] == "/meetings/" {
			templateName = "meeting_detail.html"
		} else {
			ctx.Redirect(http.StatusFound, "/login")
			return
		}
	}

	ctx.HTML(http.StatusOK, templateName, gin.H{
		"title": "Sistem Notulensi",
	})
}

// StaticPage serves frontend static assets
func (fc *FrontendController) StaticPage(ctx *gin.Context) {
	file := ctx.Param("file")
	if file == "" {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}
	ctx.File(filepath.Join("internal/templates", file))
}