package controllers

import (
	"encoding/json"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/timendus/pixelbox/models"
	"github.com/timendus/pixelbox/protocol"
	"github.com/timendus/pixelbox/server"
	xdraw "golang.org/x/image/draw"
)

func init() {
	router := http.NewServeMux()
	router.HandleFunc("GET /syncTime", syncTime)
	router.HandleFunc("POST /preview", preview)
	router.HandleFunc("POST /image", showImage)
	router.HandleFunc("POST /gif", showGif)
	server.RegisterRouter("/apply", router)
}

func preview(res http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(res, req.Body, 1<<20) // 1 MB
	defer req.Body.Close()

	var scene models.Scene
	if err := json.NewDecoder(req.Body).Decode(&scene); err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	message, err := scene.ToMessage()
	if err != nil {
		http.Error(res, "could not create message from scene: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = server.GetConnection().Send(message)
	if err != nil {
		http.Error(res, "could not apply scene: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(message)

	res.WriteHeader(http.StatusOK)
}

func syncTime(res http.ResponseWriter, req *http.Request) {
	err := server.GetConnection().Send(protocol.SetTime(time.Now()))
	if err != nil {
		http.Error(res, "could not sync time: "+err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func showImage(res http.ResponseWriter, req *http.Request) {
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

	err = server.GetConnection().Send(message)
	if err != nil {
		log.Println("could not send message:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Println("Outgoing:", message)
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func showGif(res http.ResponseWriter, req *http.Request) {
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

	err = server.GetConnection().Send(message)
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
