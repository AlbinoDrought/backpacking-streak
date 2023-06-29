package main

import (
	"embed"
	"fmt"
	"image"
	"log"
	"math/rand"
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
	looseSleep(2*time.Second, 500*time.Millisecond)
}

func main() {
	for {
		recoverMain()
		time.Sleep(time.Minute)
	}
}

func recoverMain() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	innerMain()
}

func innerMain() {
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

	nextRunTime := time.Now()

	day := 0
	firstRun := true

	for {
		if time.Now().Before(nextRunTime) {
			log.Println("i hunger")
			time.Sleep(45 * time.Minute)
			continue
		}

		if !firstRun {
			time.Sleep(30 * time.Second)
		}
		firstRun = false

		log.Println("wroooooarrr!")
		robotgo.KeyPress("f5")
		var camPerms []gcv.Result
		var takePhotos []gcv.Result
		attempts := 0
		for attempts < 24 {
			attempts++
			time.Sleep(10 * time.Second)
			log.Println("looking at the screen...")
			screen = robotgo.CaptureImg()
			camPerms = gcv.FindAllImg(imgCamPerm, screen)
			if len(camPerms) < 1 {
				log.Println("no cam perms found?")
				continue
			}
			takePhotos = gcv.FindAllImg(imgTakePhoto, screen)
			if len(takePhotos) < 1 {
				log.Println("no take photo btn found?")
				continue
			}
			break
		}
		camPerm := camPerms[0]
		takePhoto := takePhotos[0]
		log.Println("enabling camera perms")
		moveClick(camPerm.Middle.X, camPerm.Middle.Y)
		looseSleep(2*time.Second, time.Second)

		log.Println("taking a photo")
		moveClick(takePhoto.Middle.X, takePhoto.Middle.Y)
		looseSleep(2*time.Second, time.Second)

		screen = robotgo.CaptureImg()
		sendTos := gcv.FindAllImg(imgSendTo, screen)
		if len(sendTos) < 1 {
			log.Println("no send to button?")
			continue
		}

		log.Println("writing a cool message")
		moveClick(camPerm.Middle.X, camPerm.Middle.Y)
		day++
		robotgo.TypeStrDelay(fmt.Sprintf("sean is in forest (day %v), here is frog", day), 200)
		looseSleep(2*time.Second, time.Second)

		sendTo := sendTos[0]
		log.Println("pressing send to")
		moveClick(sendTo.Middle.X, sendTo.Middle.Y)

		clickBubbleLoops := 0
		tryClickBubble := true
		fireless := false
		for tryClickBubble && clickBubbleLoops < 20 {
			clickBubbleLoops++
			tryClickBubble = false

			screen = robotgo.CaptureImg(camPerm.Middle.X-300, 0, 600, camPerm.Middle.Y+300)
			fires := gcv.FindAllImg(imgFireEmoji, screen)
			if len(fires) < 1 {
				log.Println("no fire emojis found - all streaks dead? :(")
				fireless = true
				break
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
		if fireless {
			continue
		}

		log.Println("clicked all found streaks")

		screen = robotgo.CaptureImg()
		sends := gcv.FindAllImg(imgSend, screen)
		if len(sends) < 1 {
			log.Println("no send buttons found?")
			continue
		}
		send := sends[0]
		moveClick(send.Middle.X, send.Middle.Y)

		nextRunTime = nextRunTime.Add(20 * time.Hour)
	}
}
