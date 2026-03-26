package person

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"partialupdate/internal/database"
	"partialupdate/internal/models"
	"partialupdate/internal/parsers/grpc"

	"github.com/go-chi/chi/v5"
)

func CreatePerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p models.Person
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "failed to decode body", http.StatusBadRequest)
			return
		}

		slog.Info("request", "body", p)
		err = store.Create(&p)
		if err != nil {
			http.Error(w, "`id` failed to create", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", r.URL.Path+"/"+p.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode("")
	}
}

func ListPersons(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		persons, err := store.List()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(persons)
	}
}

func GetPerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		p, ok, _ := store.Get(id)
		if !ok {
			http.Error(w, "person not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}

func UpdatePerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var p models.Person
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		// Enforce that the URL id and payload id match (or you could ignore payload.id)
		if p.ID != "" && p.ID != id {
			http.Error(w, "payload id does not match URL id", http.StatusBadRequest)
			return
		}
		p.ID = id // make sure the stored record keeps the correct id

		if updated, err := store.Update(id, p); !updated {
			http.Error(w, "person not found. "+err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}

// ---------------------------------------------------------------------
// PartialUpdate (PATCH) – empty stub
// ---------------------------------------------------------------------

// PatchPerson (a.k.a. PartialUpdate) currently does nothing.
// The signature mirrors the other handlers so you can drop in real logic later.
func PatchPersonGRPC(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p grpc.Request
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "failed to decode body", http.StatusBadRequest)
			return
		}

		// TODO: implement field‑wise merge (e.g. only change Name if present)
		// For now we simply return “Not Implemented”.
		http.Error(w, "partial update not implemented yet", http.StatusNotImplemented)
	}
}

func PatchPersonPointer(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p models.PersonPatch
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "failed to decode body", http.StatusBadRequest)
			return
		}

		http.Error(w, "partial update not implemented yet", http.StatusNotImplemented)
	}
}

func DeletePerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deleted, err := store.Delete(id); !deleted {
			http.Error(w, "person not found. "+err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
