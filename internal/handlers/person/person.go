package person

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"partialupdate/internal/database"
	"partialupdate/internal/models"

	"github.com/go-chi/chi/v5"
)

func CreatePerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p models.Resource
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, "failed to decode body. "+err.Error(), http.StatusBadRequest)
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
		var p models.Resource
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		if updated, err := store.Update(id, p); !updated {
			http.Error(w, "person not found. "+err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}

// ---------------------------------------------------------------------
// PartialUpdate (PATCH)
// ---------------------------------------------------------------------
func PatchPerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented", http.StatusInternalServerError)
		// err := json.NewDecoder(r.Body).Decode(&p)
		// if err != nil {
		// 	http.Error(w, "failed to decode body", http.StatusBadRequest)
		// 	return
		// }
		// _ = chi.URLParam(r, "id")

		// status, err := store.Patch(id, p)
		// if err != nil {
		// 	switch {
		// 	case errors.Is(err, unset.ErrNoValidFields):
		// 		http.Error(w, err.Error(), http.StatusBadRequest)
		// 		return
		// 	case errors.Is(err, unset.ErrNoResourceInRequest):
		// 		http.Error(w, err.Error(), http.StatusBadRequest)
		// 		return
		// 	}
		// 	http.Error(w, "failed to decode body", http.StatusInternalServerError)
		// 	return
		// }
		//
		// if status {
		// 	w.WriteHeader(http.StatusAccepted)
		// 	return
		// }

		// w.WriteHeader(http.StatusInternalServerError)
	}
}

func DeletePerson(store *database.PersonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deleted, err := store.Delete(id); !deleted {
			if err != nil {
				http.Error(w, "person not found. "+err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, "person not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
