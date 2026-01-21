package models

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/timendus/pixelbox/protocol"
)

/* Model definitions */

type Scene struct {
	Uuid         uuid.UUID `json:"uuid"`
	Message      []byte    `json:"message"`
	MessageDirty bool      `json:"messageDirty"`

	Name             string      `json:"name"`
	Id               string      `json:"id"`
	ChangeBrightness bool        `json:"changeBrightness"`
	Brightness       *int        `json:"brightness"`
	ChangeVolume     bool        `json:"changeVolume"`
	Volume           *int        `json:"volume"`
	SceneType        string      `json:"sceneType"`
	Clock            Clock       `json:"clock"`
	Weather          Weather     `json:"weather"`
	Temperature      Temperature `json:"temperature"`
	Calendar         Calendar    `json:"calendar"`
	Light            Light       `json:"light"`
	Effect           Effect      `json:"effect"`
	Image            Image       `json:"image"`
	Animation        Animation   `json:"animation"`
}

type Clock struct {
	Enabled bool   `json:"enabled"`
	CType   string `json:"type"`
	Color   string `json:"color"`
}

type Weather struct {
	Enabled bool   `json:"enabled"`
	WType   string `json:"type"`
}

type Temperature struct {
	Enabled     bool `json:"enabled"`
	Temperature *int `json:"temperature"`
}

type Calendar struct {
	Enabled bool `json:"enabled"`
}

type Light struct {
	LType string `json:"type"`
	Color string `json:"color"`
}

type Effect struct {
	EType             string `json:"type"`
	VJType            *int   `json:"vjType"`
	VisualisationType *int   `json:"visualisationType"`
	ScoreRedPlayer    *int   `json:"scoreRedPlayer"`
	ScoreBluePlayer   *int   `json:"scoreBluePlayer"`
}

type Image struct {
	Pixels []int `json:"pixels"`
}

type Animation struct {
	Frames []Frame `json:"frames"`
}

type Frame struct {
	Duration int   `json:"duration"`
	Pixels   []int `json:"pixels"`
}

/* Scene storage stuff */

var scenes = []*Scene{}
var dir = "scenes"

func init() {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal("Can't open scene directory for reading")
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			log.Println("Can't read JSON file " + path)
			continue
		}

		var scene Scene
		if err := json.NewDecoder(f).Decode(&scene); err != nil {
			f.Close()
			log.Println("Can't decode JSON file "+path, err)
			continue
		}
		f.Close()

		scenes = append(scenes, &scene)
	}

	log.Printf("Loaded %d scenes from file\n", len(scenes))
}

func AllScenes() []*Scene {
	return scenes
}

func FindSceneByUUID(uuid uuid.UUID) (*Scene, error) {
	for _, s := range scenes {
		if s.Uuid == uuid {
			return s, nil
		}
	}
	return nil, fmt.Errorf("scene not found")
}

func FindSceneById(id string) (*Scene, error) {
	for _, s := range scenes {
		if s.Id == id {
			return s, nil
		}
	}
	return nil, fmt.Errorf("scene not found")
}

func (scene *Scene) Create() error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	scene.Uuid = id
	err = scene.writeToFile()
	if err != nil {
		return err
	}
	scenes = append(scenes, scene)
	return nil
}

func (scene *Scene) Update(newScene *Scene) error {
	scene.Name = newScene.Name
	scene.Id = newScene.Id
	scene.MessageDirty = true

	scene.ChangeBrightness = newScene.ChangeBrightness
	scene.Brightness = newScene.Brightness
	scene.ChangeVolume = newScene.ChangeVolume
	scene.Volume = newScene.Volume

	scene.SceneType = newScene.SceneType
	scene.Clock = newScene.Clock
	scene.Weather = newScene.Weather
	scene.Temperature = newScene.Temperature
	scene.Calendar = newScene.Calendar
	scene.Light = newScene.Light
	scene.Effect = newScene.Effect

	scene.Image = newScene.Image
	scene.Animation = newScene.Animation

	return scene.writeToFile()
}

func (scene *Scene) Delete() error {
	err := scene.deleteFile()
	if err != nil {
		return err
	}
	for i, s := range scenes {
		if s.Uuid == scene.Uuid {
			scenes = append(scenes[:i], scenes[i+1:]...)
			return nil
		}
	}
	return nil
}

func (scene *Scene) writeToFile() error {
	path := filepath.Join(dir, scene.Uuid.String()+".json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	return enc.Encode(scene)
}

func (scene *Scene) deleteFile() error {
	path := filepath.Join(dir, scene.Uuid.String()+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

/* Conversion to Divoom Timebox Evo message stuff */

func (scene *Scene) GetMessage() ([]byte, error) {
	if scene.MessageDirty {
		var err error
		scene.Message, err = scene.ToMessage()
		if err != nil {
			return nil, err
		}
		scene.MessageDirty = false
	}
	return scene.Message, nil
}

func (scene *Scene) ToMessage() ([]byte, error) {
	result := make([]byte, 0)

	// Brightness setting
	if scene.SceneType != "light" && scene.ChangeBrightness && scene.Brightness != nil {
		msg, err := protocol.SetBrightness(*scene.Brightness)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)
	}

	// Volume setting
	if scene.ChangeVolume && scene.Volume != nil {
		msg, err := protocol.SetVolume(*scene.Volume)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)
	}

	switch scene.SceneType {
	case "clock":
		msg, err := protocol.ShowClock(
			protocol.ClockType(scene.Clock.CType),
			scene.Clock.Enabled,
			scene.Weather.Enabled,
			scene.Temperature.Enabled,
			scene.Calendar.Enabled,
			protocol.ColorFromHex(scene.Clock.Color),
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

		if scene.Weather.Enabled || scene.Temperature.Enabled {
			temp := 0
			if scene.Temperature.Temperature != nil {
				temp = *scene.Temperature.Temperature
			}
			msg, err := protocol.SetWeather(temp, protocol.WeatherType(scene.Weather.WType))
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)
		}

	case "light":
		if scene.Brightness == nil {
			return nil, fmt.Errorf("brightness needs to be set to be able to show a light")
		}
		msg, err := protocol.ShowLight(
			protocol.LightType(scene.Light.LType),
			protocol.ColorFromHex(scene.Light.Color),
			*scene.Brightness,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

	case "effects":
		switch scene.Effect.EType {
		case "CLOUD":
			result = append(result, protocol.ShowCloud()...)

		case "VJ":
			if scene.Effect.VJType == nil {
				return nil, fmt.Errorf("VJ type required for showing a VJ effect")
			}
			msg, err := protocol.ShowVJEffect(*scene.Effect.VJType)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)

		case "VISUALISATION":
			if scene.Effect.VisualisationType == nil {
				return nil, fmt.Errorf("visualisation type required for showing a visualisation")
			}
			msg, err := protocol.ShowVisualisation(*scene.Effect.VisualisationType)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)

		case "SCOREBOARD":
			if scene.Effect.ScoreBluePlayer == nil || scene.Effect.ScoreRedPlayer == nil {
				return nil, fmt.Errorf("player scores required for showing the score board")
			}
			msg, err := protocol.ShowScoreBoard(*scene.Effect.ScoreRedPlayer, *scene.Effect.ScoreBluePlayer)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)
		}

	case "image":
		img := image.NewRGBA(image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: 16, Y: 16},
		})
		for i, v := range scene.Image.Pixels {
			img.Pix[i] = byte(v)
		}
		msg, err := protocol.ShowImage(img)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

	case "animation":
		frames := make([]*image.RGBA, 0)
		durations := make([]int, 0)
		for _, f := range scene.Animation.Frames {
			img := image.NewRGBA(image.Rectangle{
				Min: image.Point{X: 0, Y: 0},
				Max: image.Point{X: 16, Y: 16},
			})
			for i, v := range f.Pixels {
				img.Pix[i] = byte(v)
			}
			frames = append(frames, img)
			durations = append(durations, f.Duration)
		}
		msg, err := protocol.ShowAnimation(frames, durations)
		if err != nil {
			return nil, err
		}
		result = append(result, msg...)

	}

	return result, nil
}
