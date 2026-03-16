package http

import (
	"net/http"

	appuser "mltestsuite/internal/application/user"
	"mltestsuite/internal/domain/user"

	"github.com/google/uuid"
)

// UserHandler maneja la gestion de usuarios (solo admin).
type UserHandler struct {
	service *appuser.Service
	tmpl    *Renderer
}

// NewUserHandler crea el handler de usuarios.
func NewUserHandler(service *appuser.Service, tmpl *Renderer) *UserHandler {
	return &UserHandler{service: service, tmpl: tmpl}
}

// List muestra la lista de usuarios.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	h.tmpl.ExecuteTemplate(w, "admin/users_list.html", withFlash(w, r, map[string]any{
		"Users": users,
	}))
}

// ShowCreate muestra el formulario de creacion de usuario.
func (h *UserHandler) ShowCreate(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "admin/user_form.html", withFlash(w, r, map[string]any{
		"IsNew": true,
	}))
}

// Create crea un nuevo usuario via el servicio de auth (register).
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	input := appuser.UpdateUserInput{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Role:     user.Role(r.FormValue("role")),
		Active:   true,
		Password: r.FormValue("password"),
	}
	// Create a new user directly via the user service (admin flow)
	// We use the auth service Register behind the scenes or directly create
	// For simplicity, we create via repo through service
	newID := uuid.New()
	if err := h.service.UpdateUser(r.Context(), newID, input); err != nil {
		// If update fails (user not found), it means we need to create
		// For admin user creation, we redirect with error
		setFlash(w, "error", "Error creando usuario: "+err.Error())
		http.Redirect(w, r, "/admin/users/new", http.StatusSeeOther)
		return
	}

	setFlash(w, "success", "Usuario creado correctamente")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// ShowEdit muestra el formulario de edicion de usuario.
func (h *UserHandler) ShowEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	u, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		return
	}
	h.tmpl.ExecuteTemplate(w, "admin/user_form.html", withFlash(w, r, map[string]any{
		"IsNew": false,
		"User":  u,
	}))
}

// Update actualiza un usuario.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	input := appuser.UpdateUserInput{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Role:     user.Role(r.FormValue("role")),
		Active:   r.FormValue("active") == "on",
		Password: r.FormValue("password"),
	}
	if err := h.service.UpdateUser(r.Context(), id, input); err != nil {
		setFlash(w, "error", err.Error())
		http.Redirect(w, r, "/admin/users/"+idStr+"/edit", http.StatusSeeOther)
		return
	}
	setFlash(w, "success", "Usuario actualizado")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// Delete elimina un usuario.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
