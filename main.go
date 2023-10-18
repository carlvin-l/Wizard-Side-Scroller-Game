package main

import (
	"fmt"
	"github.com/co0p/tankism/lib/collision"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"math/rand"
	"os"
)

// main struct
type arcade struct {
	player          Sprite
	enemy           Sprite
	enemies         []Sprite
	fireballs       []Sprite
	backgroundXView int
	score           int
	speed           int
	counter         int
	background      *ebiten.Image
	audioContext    *audio.Context
	soundPlayer     *audio.Player
	hitPlayer       *audio.Player
}

// Sprite side struct
type Sprite struct {
	pict   *ebiten.Image
	xLoc   int
	yLoc   int
	active bool
}

// main function
func main() {
	ebiten.SetWindowSize(1000, 750)
	ebiten.SetWindowTitle("The Farmer Who Was Summoned To Another World With an OP Skill")
	playerPict, _, err := ebitenutil.NewImageFromFile("wiz.png")
	if err != nil {
		fmt.Println("Error loading image", err)
	}
	badGuyPict, _, err := ebitenutil.NewImageFromFile("enemy.png")
	if err != nil {
		fmt.Println("Error loading image", err)
	}
	backgroundPict, _, err := ebitenutil.NewImageFromFile("night.png")
	if err != nil {
		fmt.Println("Unable to load background image:", err)
	}
	const SoundSampleRate = 48000
	soundContext := audio.NewContext(SoundSampleRate)
	ourGame := arcade{
		background:   backgroundPict,
		player:       Sprite{playerPict, 0, 300, true},
		enemy:        Sprite{badGuyPict, 1000, rand.Intn(750), false},
		enemies:      make([]Sprite, 0),
		audioContext: soundContext,
		soundPlayer:  LoadWav("fireball.wav", soundContext),
		hitPlayer:    LoadWav("impact.wav", soundContext),
		counter:      20,
		score:        0,
	}
	// creates a group of enemies
	for i := 0; i < 3; i++ {
		enemy := Sprite{badGuyPict, 1000, rand.Intn(750), false}
		ourGame.enemies = append(ourGame.enemies, enemy)
	}
	err = ebiten.RunGame(&ourGame)
	if err != nil {
		fmt.Println("Failed to run game", err)
	}
}

// Update function used to update the game as it is running
func (game *arcade) Update() error {
	backgroundWidth := game.background.Bounds().Dx()
	maxX := backgroundWidth * 2
	game.backgroundXView -= 4
	game.backgroundXView %= maxX
	// enemies continuous movement along with update if enemy makes it across the screen
	for i := range game.enemies {
		game.enemies[i].xLoc -= 4
		if game.enemies[i].xLoc < -100 {
			game.enemies[i].xLoc = 1000
			game.enemies[i].yLoc = rand.Intn(750)
			game.score--
		}
	}
	// controls for the player's movement
	game.speed = 2
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		game.player.yLoc += game.speed - 7
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		game.player.yLoc += game.speed + 3
	}
	// grabs the fire image from the project folder
	shotPict, _, err := ebitenutil.NewImageFromFile("fire.png")
	if err != nil {
		fmt.Println("Unable to load background image:", err)
	}
	// updates the fireball to spawn at wizard's staff
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		newFireBall := Sprite{
			pict: shotPict,
			xLoc: game.player.xLoc + 20,
			yLoc: game.player.yLoc - 9,
		}
		// plays a sound when a fireball is shot
		game.fireballs = append(game.fireballs, newFireBall)
		err := game.soundPlayer.Rewind()
		if err != nil {
			return err
		}
		game.soundPlayer.Play()
		game.counter = 0
	}
	// creates a splice for the enemies hit
	var enemyHit []int
	// updates the fireball to move across the screen
	for i := 0; i < len(game.fireballs); i++ {
		game.fireballs[i].xLoc += 5
		// collision checking with enemies
		for j := 0; j < len(game.enemies); j++ {
			badMan := collision.BoundingBox{
				X:      float64(game.enemies[j].xLoc),
				Y:      float64(game.enemies[j].yLoc),
				Width:  float64(game.enemies[j].pict.Bounds().Dx()),
				Height: float64(game.enemies[j].pict.Bounds().Dy()),
			}
			spellCast := collision.BoundingBox{
				X:      float64(game.fireballs[i].xLoc),
				Y:      float64(game.fireballs[i].yLoc),
				Width:  float64(game.fireballs[i].pict.Bounds().Dx()),
				Height: float64(game.fireballs[i].pict.Bounds().Dy()),
			}
			// resetting the enemies position and playing sound when 'hit'
			if collision.AABBCollision(spellCast, badMan) {
				enemyHit = append(enemyHit, i)
				game.enemies[j].xLoc = 1000
				game.enemies[j].yLoc = rand.Intn(700)
				err := game.hitPlayer.Rewind()
				if err != nil {
					return err
				}
				game.hitPlayer.Play()
				game.score += 2
			}
		}
	}
	// removes the fireball from screen after collision
	for i := len(enemyHit) - 1; i >= 0; i-- {
		index := enemyHit[i]
		game.fireballs = append(game.fireballs[:index], game.fireballs[index+1:]...)
	}
	return nil
}

// Draw function used to draw everything on the screen
func (game *arcade) Draw(screen *ebiten.Image) {
	// draws the background for the game on screen
	drawOps := ebiten.DrawImageOptions{}
	const repeat = 5
	backgroundWidth := game.background.Bounds().Dx()
	for count := 0; count < repeat; count += 1 {
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(backgroundWidth*count),
			float64(-250))
		drawOps.GeoM.Translate(float64(game.backgroundXView), 0)
		screen.DrawImage(game.background, &drawOps)
	}
	// displays the score at the top left
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", game.score), 10, 10)
	// draws the player on screen
	drawOps.GeoM.Reset()
	drawOps.GeoM.Translate(float64(game.player.xLoc), float64(game.player.yLoc))
	screen.DrawImage(game.player.pict, &drawOps)
	// draws each fireball on screen
	for i := 0; i < len(game.fireballs); i++ {
		fireball := game.fireballs[i]
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(fireball.xLoc), float64(fireball.yLoc))
		screen.DrawImage(fireball.pict, &drawOps)
	}
	// draws each enemy on screen
	for i := 0; i < len(game.enemies); i++ {
		enemy := game.enemies[i]
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(enemy.xLoc), float64(enemy.yLoc))
		screen.DrawImage(enemy.pict, &drawOps)
	}
}

// LoadWav function used to load the Wav files
func LoadWav(name string, context *audio.Context) *audio.Player {
	soundFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Error Loading sound: ", err)
	}
	soundEffect, err := wav.DecodeWithoutResampling(soundFile)
	if err != nil {
		fmt.Println("Error interpreting sound file: ", err)
	}
	soundPlayer, err := context.NewPlayer(soundEffect)
	if err != nil {
		fmt.Println("Couldn't create sound player: ", err)
	}
	_, err = context.NewPlayer(soundEffect)
	if err != nil {
		fmt.Println("Couldn't create sound player: ", err)
	}
	return soundPlayer

}

// Layout function used to make a layout for the game
func (game arcade) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
