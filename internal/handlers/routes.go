package handlers

import (
	"io/fs"
	"net/http"
)

type Router struct {
	mux       *http.ServeMux
	dashboard *DashboardHandler
	payment   *PaymentHandler
	staticFS  fs.FS
}

func NewRouter(d *DashboardHandler, p *PaymentHandler, staticFS fs.FS) *Router {
	r := &Router{
		mux:       http.NewServeMux(),
		dashboard: d,
		payment:   p,
		staticFS:  staticFS,
	}
	r.registerRoutes()
	return r
}

func (r *Router) registerRoutes() {
	// Static assets
	if r.staticFS != nil {
		r.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(r.staticFS))))
	}

	// Public landing & parent payment links
	r.mux.HandleFunc("GET /{$}", r.payment.ServeLandingPage)
	r.mux.HandleFunc("GET /pay/{slug}", r.payment.ServePayPage)
	r.mux.HandleFunc("POST /api/pay/stk-push", r.payment.InitiateSTKPush)
	r.mux.HandleFunc("POST /api/pay/confirm", r.payment.ConfirmPayment)
	r.mux.HandleFunc("GET /receipt/{id}", r.payment.ServeReceiptPage)

	// School Admin Dashboard
	r.mux.HandleFunc("GET /admin", r.dashboard.ServeDashboard)
	r.mux.HandleFunc("POST /api/students", r.dashboard.AddStudent)
	r.mux.HandleFunc("POST /api/students/bulk", r.dashboard.BulkImport)
	r.mux.HandleFunc("DELETE /api/students/{id}", r.dashboard.DeleteStudent)
	r.mux.HandleFunc("POST /api/students/{id}/delete", r.dashboard.DeleteStudent)
	r.mux.HandleFunc("POST /api/students/{id}/remind", r.dashboard.SendReminder)
	r.mux.HandleFunc("POST /api/remind-all", r.dashboard.RemindAllOwed)
	r.mux.HandleFunc("GET /admin/export.csv", r.dashboard.ExportCSV)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Enable CORS & Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	r.mux.ServeHTTP(w, req)
}
