package server

import (
	"errors"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"net/http"
)

func (s *Server) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var userRegister domain.UserRegister
	if err := s.readJSON(w, r, &userRegister); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	if err := s.Facade.RegisterUser(r.Context(), &userRegister); err != nil {
		var ev *domain.ErrValidation
		switch {
		case errors.As(err, &ev):
			s.failedValidationResponse(w, r, ev.Errors)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetByUniqueFieldHandler(w http.ResponseWriter, r *http.Request) {
	fieldValue := r.PathValue("field")
	user, err := s.Facade.GetByUniqueField(r.Context(), fieldValue)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRecordNotFound):
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
	if err = s.writeJSON(w, envelop{"user": user}, http.StatusOK, nil); err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userUpdate domain.UserUpdate
	if err := s.readJSON(w, r, &userUpdate); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	if err := s.Facade.UpdateUser(r.Context(), &userUpdate); err != nil {
		var ev *domain.ErrValidation
		switch {
		case errors.As(err, &ev):
			s.failedValidationResponse(w, r, ev.Errors)
		case errors.Is(err, domain.ErrEditConflict):
			s.editConflictResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
}

func (s *Server) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var token struct {
		OTP string `json:"otp"`
	}
	if err := s.readJSON(w, r, &token); err != nil {
		s.badRequestResponse(w, r, err)
		return
	}
	if err := s.Facade.ActivateUser(r.Context(), token.OTP); err != nil {
		var ev *domain.ErrValidation
		switch {
		case errors.As(err, &ev):
			s.failedValidationResponse(w, r, ev.Errors)
		case errors.Is(err, domain.ErrAlreadyActive):
			s.alreadyActivatedResponse(w, r)
		case errors.Is(err, domain.ErrEditConflict):
			s.editConflictResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}
}
