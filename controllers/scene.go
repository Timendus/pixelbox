package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/timendus/pixelbox/models"
	"github.com/timendus/pixelbox/server"
)

func init() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", sceneList)
	router.HandleFunc("PUT /", newScene)
	router.HandleFunc("GET /{id}", getScene)
	router.HandleFunc("DELETE /{id}", deleteScene)
	router.HandleFunc("POST /{id}", updateScene)
	router.HandleFunc("GET /{id}/apply", applyScene)
	server.RegisterRouter("/scene", router)
}

func sceneList(res http.ResponseWriter, req *http.Request) {
	scenes := models.AllScenes()
	json.NewEncoder(res).Encode(scenes)
}

func newScene(res http.ResponseWriter, req *http.Request) {
	brightness := 100
	volume := 16
	temperature := 20
	scene := models.Scene{
		Name:       "New Scene",
		Id:         "new-scene",
		Brightness: &brightness,
		Volume:     &volume,
		SceneType:  "clock",
		Clock: models.Clock{
			Enabled: true,
			CType:   "FULL_SCREEN",
			Color:   "#FF0000",
		},
		Weather: models.Weather{
			WType: "OUTDOOR_VERY_LIGHT_CLOUDS",
		},
		Temperature: models.Temperature{
			Temperature: &temperature,
		},
		Light: models.Light{
			LType: "PLAIN",
			Color: "#FFFF00",
		},
		Effect: models.Effect{
			EType: "CLOUD",
		},
	}
	err := scene.Create()
	if err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, "/scene/", http.StatusSeeOther)
}

func getScene(res http.ResponseWriter, req *http.Request) {
	uuid, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(res, "invalid ID requested", http.StatusBadRequest)
		return
	}
	scene, err := models.FindSceneByUUID(uuid)
	if err != nil {
		http.Error(res, "Could not find model with given ID", http.StatusNotFound)
		return
	}
	json.NewEncoder(res).Encode(scene)
}

func deleteScene(res http.ResponseWriter, req *http.Request) {
	uuid, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(res, "invalid ID requested", http.StatusBadRequest)
		return
	}
	scene, err := models.FindSceneByUUID(uuid)
	if err != nil {
		http.Error(res, "Could not find model with given ID", http.StatusNotFound)
		return
	}
	err = scene.Delete()
	if err != nil {
		http.Error(res, "Could not delete model", http.StatusInternalServerError)
		return
	}
	http.Redirect(res, req, "/scene/", http.StatusSeeOther)
}

func updateScene(res http.ResponseWriter, req *http.Request) {
	uuid, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		http.Error(res, "invalid ID requested", http.StatusBadRequest)
		return
	}

	scene, err := models.FindSceneByUUID(uuid)
	if err != nil {
		http.Error(res, "Could not find model with given ID", http.StatusNotFound)
		return
	}

	req.Body = http.MaxBytesReader(res, req.Body, 1<<20) // 1 MB
	defer req.Body.Close()

	var newScene models.Scene
	if err := json.NewDecoder(req.Body).Decode(&newScene); err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = scene.Update(&newScene)
	if err != nil {
		http.Error(res, "Could not update model", http.StatusInternalServerError)
		return
	}

	http.Redirect(res, req, "/scene/", http.StatusSeeOther)
}

func applyScene(res http.ResponseWriter, req *http.Request) {
	scene, err := models.FindSceneById(req.PathValue("id"))
	if err != nil {
		uuid, err := uuid.Parse(req.PathValue("id"))
		if err != nil {
			http.Error(res, "invalid ID requested", http.StatusBadRequest)
			return
		}
		scene, err = models.FindSceneByUUID(uuid)
		if err != nil {
			http.Error(res, "Could not find model with given ID", http.StatusNotFound)
			return
		}
	}

	message, err := scene.GetMessage()
	if err != nil {
		http.Error(res, "could not create message from scene: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = server.GetConnection().Send(message)
	if err != nil {
		http.Error(res, "could not apply scene: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
