package v1

import (
	"partialupdate/internal/handlers/person"
	"partialupdate/internal/routes"

	"github.com/go-chi/chi/v5"
)

func SetupV1Routes(router *routes.Router) {

	router.Router.Route("/v1/person", func(v1person chi.Router) {
		v1person.Post("/", person.CreatePerson(router.Store)) // POST   /v1/person
		v1person.Get("/", person.ListPersons(router.Store))   // GET    /v1/person
		v1person.Route("/{id}", func(personid chi.Router) {
			personid.Get("/", person.GetPerson(router.Store))         // GET    /v1/person/{id}
			personid.Put("/", person.UpdatePerson(router.Store))      // PUT    /v1/person/{id}
			personid.Patch("/", person.PatchPersonGRPC(router.Store)) // PATCH  /v1/person/{id}
			// personid.Patch("/", person.PatchPerson(router.Store))   // PATCH  /v1/person/{id}
			personid.Delete("/", person.DeletePerson(router.Store)) // DELETE /v1/person/{id}
		})
	})
}
