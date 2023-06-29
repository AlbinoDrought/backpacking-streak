package main

import (
	"embed"
	"image"
	"log"
	"math/rand"
	"os"
	"path"
	"sort"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/gcv"

	_ "image/png"
)

//go:embed steps/*.png
var steps embed.FS

func loadStep(p string) (image.Image, error) {
	handle, err := steps.Open(path.Join("steps/", p))
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	img, _, err := image.Decode(handle)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func looseSleep(about time.Duration, giveOrTake time.Duration) {
	give := rand.Int()%2 == 0
	mod := time.Millisecond * time.Duration(rand.Int63n(giveOrTake.Milliseconds()))
	if give {
		time.Sleep(about + mod)
	} else {
		time.Sleep(about - mod)
	}
}

func moveClick(x int, y int) {
	robotgo.MoveSmooth(x, y)
	looseSleep(1*time.Second, 500*time.Millisecond)
	robotgo.Click()
	looseSleep(3*time.Second, 500*time.Millisecond)
}

func main() {
	imgCamPerm, err := loadStep("1_camperm.png")
	if err != nil {
		panic(err)
	}
	imgTakePhoto, err := loadStep("2_takephoto_oncamperm.png")
	if err != nil {
		panic(err)
	}
	imgSendTo, err := loadStep("3_sendto.png")
	if err != nil {
		panic(err)
	}
	imgFireEmoji, err := loadStep("4_fireemoji.png")
	if err != nil {
		panic(err)
	}
	imgUnselected, err := loadStep("5_unselected.png")
	if err != nil {
		panic(err)
	}
	imgSelected, err := loadStep("6_selected.png")
	if err != nil {
		panic(err)
	}
	imgSend, err := loadStep("7_send.png")
	if err != nil {
		panic(err)
	}

	var screen image.Image

	log.Println("giving you time to focus chromium...")
	time.Sleep(2 * time.Second)

	// robotgo.KeyPress("f5")
	// time.Sleep(time.Minute)

	log.Println("looking at the screen...")
	screen = robotgo.CaptureImg()
	camPerms := gcv.FindAllImg(imgCamPerm, screen)
	if len(camPerms) < 1 {
		panic("todo: wait longer here")
	}
	takePhotos := gcv.FindAllImg(imgTakePhoto, screen)
	if len(takePhotos) < 1 {
		panic("todo: wait longer here 2")
	}
	camPerm := camPerms[0]
	takePhoto := takePhotos[0]
	log.Println("enabling camera perms")
	moveClick(camPerm.Middle.X, camPerm.Middle.Y)

	log.Println("taking a photo")
	moveClick(takePhoto.Middle.X, takePhoto.Middle.Y)

	screen = robotgo.CaptureImg()
	sendTos := gcv.FindAllImg(imgSendTo, screen)
	if len(sendTos) < 1 {
		panic("todo: retry, or F5, or whatever here")
	}
	sendTo := sendTos[0]
	log.Println("pressing send to")
	moveClick(sendTo.Middle.X, sendTo.Middle.Y)

	clickBubbleLoops := 0
	tryClickBubble := true
	for tryClickBubble && clickBubbleLoops < 20 {
		clickBubbleLoops++
		tryClickBubble = false

		screen = robotgo.CaptureImg(camPerm.Middle.X-300, 0, 600, camPerm.Middle.Y+300)
		fires := gcv.FindAllImg(imgFireEmoji, screen)
		if len(fires) < 1 {
			panic("todo: retry, or all streaks have died")
		}
		sort.SliceStable(fires, func(i, j int) bool {
			if fires[i].Middle.X < fires[j].Middle.X || fires[i].Middle.Y < fires[j].Middle.Y {
				return true
			}
			return false
		})
		for _, fire := range fires {
			// look for unselected bubble
			screen = robotgo.CaptureImg(fire.Middle.X+camPerm.Middle.X-300, fire.TopLeft.Y-25, 150, imgUnselected.Bounds().Dy()+50)
			unselecteds := gcv.FindAllImg(imgUnselected, screen)
			selecteds := gcv.FindAllImg(imgSelected, screen)
			if len(unselecteds) > len(selecteds) {
				log.Println("found a streak, clicking")
				moveClick(fire.Middle.X+camPerm.Middle.X-300, fire.Middle.Y)
				tryClickBubble = true
				break
			}
		}
	}

	log.Println("clicked all found streaks")
	os.Exit(0)

	if false {
		screen = robotgo.CaptureImg()
		sends := gcv.FindAllImg(imgSend, screen)
		if len(sends) < 1 {
			panic("todo: retry, or its fucked")
		}
		send := sends[0]
		moveClick(send.Middle.X, send.Middle.Y)
	}
}
