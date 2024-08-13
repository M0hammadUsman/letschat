package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type envelop map[string]any

func (*Server) writeJSON(w http.ResponseWriter, data envelop, status int, headers http.Header) error {
	jsonBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	for k, v := range headers {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)
	return nil
}

func (*Server) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeErr.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (s *Server) readInt(v url.Values, key string, defaultValue int, ev *domain.ErrValidation) int {
	str := v.Get(key)
	if str == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		ev.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

func (s *Server) readString(v url.Values, key, defaultValue string) string {
	str := v.Get(key)
	if str == "" {
		return defaultValue
	}
	return str
}
