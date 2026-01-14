package main

import (
	"image"
	"image/draw"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/timendus/pixelbox/protocol"
	xdraw "golang.org/x/image/draw"
)

var connection Connection

func main() {
	log.Println("Starting...")

	go func() {
		connection = NewConnection("11:75:58:B1:B2:15", 1, callback)
		err := connection.Connect()
		if err != nil {
			log.Println("Could not connect to the Divoom Timebox Evo:", err)
		} else {
			log.Println("Connected to Divoom Timebox Evo")
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/api", apiAction)
	http.HandleFunc("/image", imageAction)
	http.HandleFunc("/animation", gifAction)

	server := &http.Server{Addr: ":3000"}
	log.Println("Listening on port 3000")
	server.ListenAndServe()
}

func callback(message []byte) {
	commands, err := protocol.ParseIncoming(message)
	if err != nil {
		log.Println("Could not parse message:", err, message)
		return
	}
	log.Println("Parsed incoming message as:")
	for _, command := range commands {
		log.Println(" *", command)
	}
}

func apiAction(res http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	var message []byte

	if params.Has("brightness") {
		value, err := strconv.Atoi(params.Get("brightness"))
		if err != nil {
			log.Println("invalid brightness:", params.Get("brightness"))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		message, err = protocol.SetBrightness(value)
		if err != nil {
			log.Println("invalid input:", value, err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if params.Has("volume") {
		value, err := strconv.Atoi(params.Get("volume"))
		if err != nil {
			log.Println("invalid volume:", params.Get("volume"))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		message, err = protocol.SetVolume(value)
		if err != nil {
			log.Println("invalid input:", value, err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if params.Has("clock") {
		showClock := params.Get("showClock") == "1"
		showWeather := params.Get("showWeather") == "1"
		showTemperature := params.Get("showTemperature") == "1"
		showCalendar := params.Get("showCalendar") == "1"
		color := protocol.Color{255, 255, 255}
		if params.Has("clockColor") {
			parts := strings.Split(params.Get("clockColor"), ",")
			if len(parts) != 3 {
				log.Println("invalid color:", params.Get("clockColor"))
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			red, rErr := strconv.Atoi(parts[0])
			green, gErr := strconv.Atoi(parts[1])
			blue, bErr := strconv.Atoi(parts[2])
			if rErr != nil || gErr != nil || bErr != nil {
				log.Println("invalid color:", params.Get("clockColor"), rErr, gErr, bErr)
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			color = protocol.Color{byte(red), byte(green), byte(blue)}
		}
		var err error
		message, err = protocol.ShowClock(protocol.ClockType(params.Get("clock")), showClock, showWeather, showTemperature, showCalendar, color)
		if err != nil {
			log.Println("invalid input:", params.Get("clock"), err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if params.Has("syncTime") {
		message = protocol.SetTime(time.Now())
	}

	if params.Has("temperature") && params.Has("weather") {
		temprValue, err := strconv.Atoi(params.Get("temperature"))
		if err != nil {
			log.Println("invalid temperature:", params.Get("temperature"))
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		message, err = protocol.SetWeather(temprValue, protocol.WeatherType(params.Get("weather")))
		if err != nil {
			log.Println("invalid input:", temprValue, params.Get("weather"), err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if params.Has("light") {
		brightness, err := strconv.Atoi(params.Get("lightBrightness"))
		if err != nil {
			brightness = 100
		}
		color := protocol.Color{255, 255, 255}
		if params.Has("lightColor") {
			parts := strings.Split(params.Get("lightColor"), ",")
			if len(parts) != 3 {
				log.Println("invalid color:", params.Get("lightColor"))
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			red, rErr := strconv.Atoi(parts[0])
			green, gErr := strconv.Atoi(parts[1])
			blue, bErr := strconv.Atoi(parts[2])
			if rErr != nil || gErr != nil || bErr != nil {
				log.Println("invalid color:", params.Get("lightColor"), rErr, gErr, bErr)
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			color = protocol.Color{byte(red), byte(green), byte(blue)}
		}
		message, err = protocol.ShowLight(protocol.LightType(params.Get("light")), color, brightness)
		if err != nil {
			log.Println("invalid input:", params.Get("light"), color, brightness, err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if params.Has("displayOff") {
		message = protocol.DisplayOff()
	}

	err := connection.Send(message)
	if err != nil {
		log.Println("could not send message:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Println("Outgoing:", message)
	res.WriteHeader(http.StatusOK)
}

func imageAction(res http.ResponseWriter, req *http.Request) {
	// Limit size defensively (example: 10 MB)
	req.Body = http.MaxBytesReader(res, req.Body, 10<<20)

	if err := req.ParseMultipartForm(10 << 20); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(res, "invalid image", http.StatusBadRequest)
		return
	}

	message, err := protocol.ShowImage(toScaledRGBA(img))
	if err != nil {
		log.Println("could not show image:", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = connection.Send(message)
	if err != nil {
		log.Println("could not send message:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Println("Outgoing:", message)
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func gifAction(res http.ResponseWriter, req *http.Request) {
	// Limit size defensively (example: 10 MB)
	req.Body = http.MaxBytesReader(res, req.Body, 10<<20)

	if err := req.ParseMultipartForm(10 << 20); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := gif.DecodeAll(file)
	if err != nil {
		http.Error(res, "invalid animation", http.StatusBadRequest)
		return
	}

	frames := extractFullGIFFrames(img)

	for i, frame := range frames {
		frames[i] = toScaledRGBA(frame)
		img.Delay[i] *= 10 // convert to ms
	}

	message, err := protocol.ShowAnimation(frames, img.Delay)
	if err != nil {
		log.Println("could not show image:", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = connection.Send(message)
	if err != nil {
		log.Println("could not send message:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Println("Outgoing:", message)
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func extractFullGIFFrames(g *gif.GIF) []*image.RGBA {
	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)

	var frames []*image.RGBA

	for i, frame := range g.Image {
		// Save a copy of the canvas *before* drawing the frame
		prev := image.NewRGBA(bounds)
		draw.Draw(prev, bounds, canvas, image.Point{}, draw.Src)

		// Draw frame onto canvas
		draw.Draw(canvas, frame.Bounds(), frame, image.Point{}, draw.Over)

		// Capture the composited frame
		out := image.NewRGBA(bounds)
		draw.Draw(out, bounds, canvas, image.Point{}, draw.Src)
		frames = append(frames, out)

		// Handle disposal
		switch g.Disposal[i] {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			draw.Draw(canvas, bounds, prev, image.Point{}, draw.Src)
		case gif.DisposalNone:
			// keep canvas as-is
		}
	}

	return frames
}

func toScaledRGBA(img image.Image) *image.RGBA {
	const target = 16

	src := img.Bounds()
	sw, sh := src.Dx(), src.Dy()

	// Don't scale the image if it is already the right resolution
	// (Do we need this? Not for the result, but it is faster...)
	if sw == 16 && sh == 16 {
		// Fast path: already RGBA
		if rgba, ok := img.(*image.RGBA); ok {
			return rgba
		}

		rgba := image.NewRGBA(src)
		// Draw copies pixels and handles colorspace conversion
		draw.Draw(rgba, src, img, src.Min, draw.Src)

		return rgba
	}

	// Scale to fit within 16x16, preserving aspect ratio
	scale := math.Min(
		float64(target)/float64(sw),
		float64(target)/float64(sh),
	)

	dw := int(math.Round(float64(sw) * scale))
	dh := int(math.Round(float64(sh) * scale))
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}

	// 1) Resize into a temporary image of size dw x dh
	scaled := image.NewRGBA(image.Rect(0, 0, dw, dh))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), img, src, xdraw.Over, nil)

	// 2) Composite onto a 16x16 canvas, centered (letterbox)
	out := image.NewRGBA(image.Rect(0, 0, target, target))

	ox := (target - dw) / 2
	oy := (target - dh) / 2

	draw.Draw(
		out,
		image.Rect(ox, oy, ox+dw, oy+dh),
		scaled,
		image.Point{},
		draw.Over,
	)

	return out
}
