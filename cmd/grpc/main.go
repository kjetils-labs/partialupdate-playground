package main

import (
	"encoding/json"
	"log/slog"
	"partialupdate/internal/models"
	"partialupdate/internal/parsers/grpc"
)

func AttemptParse(request *grpc.Request) {
	slog.Info("testing valid fields in updateMask")
	bytes, err := request.Parse()
	if err != nil {
		slog.Error("failed parsing", "error", err)
	}
	raw, _ := json.Marshal(request.Resource)
	slog.Info("raw_data", slog.Any("data", string(raw)))
	slog.Info("parsed_data", slog.Any("data", string(bytes)))
}

func main() {
	person := models.Person{
		Name:  "Cheese Man",
		Alive: true,
		Age:   69,
	}

	request := grpc.NewRequest(person, []string{"name", "alive"})
	AttemptParse(request)

	slog.Info("testing no valid fields in updateMask")
	request = grpc.NewRequest(person, []string{"name", "alive"})
	AttemptParse(request)
}
