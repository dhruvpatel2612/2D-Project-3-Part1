package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/lafriks/go-tiled"
)

type soundDemo struct {
	audioContext *audio.Context
	soundPlayer  *audio.Player
	counter      int
} //comment for push

func (demo *soundDemo) Update() error {
	demo.counter += 1
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		demo.soundPlayer.Rewind()
		demo.soundPlayer.Play()
		demo.counter = 0
	}
	return nil
}

func LoadWav(name string, context *audio.Context) *audio.Player {
	thunderFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Error Loading sound: ", err)
	}
	thunderSound, err := wav.DecodeWithoutResampling(thunderFile)
	if err != nil {
		fmt.Println("Error interpreting sound file: ", err)
	}
	soundPlayer, err := context.NewPlayer(thunderSound)
	if err != nil {
		fmt.Println("Couldn't create sound player: ", err)
	}
	return soundPlayer
}

const (
	mapPath            = "assets/maps/map.tmx"
	nextLevelMapPath   = "assets/maps/NewLevelMap.tmx"
	enemySpriteSheet   = "assets/images/enemy.png"
	enemySpriteWidth   = 24
	enemySpriteHeight  = 30
	enemySpriteColumns = 6
	enemySpriteRows    = 4
	enemySpeed         = 1.0
	shotImageFilePath  = "assets/images/shots.png"

	towerShootInterval = 2 * time.Second // Adjust the interval as needed
	shotSpeed          = 5.0             // Adjust the shot speed as needed
	shot2ImageFilePath = "assets/images/shots2.png"
)

var towerPaths = []string{
	"assets/images/tower1.png",
	"assets/images/tower2.png",
	"assets/images/tower3.png",
	"assets/images/tower4.png",
}

var towerCosts = []int{
	300, // Tower 1 cost
	400, // Tower 2 cost
	500, // Tower 3 cost
	200, // Tower 4 cost
}
var currentMap = "assets/maps/map.tmx"

const (
	buttonSize         = 50
	buttonScale        = 0.5
	buttonMargin       = 10
	selectionAreaWidth = 100
)

type Button struct {
	Image *ebiten.Image
	X, Y  int
}

type Enemy struct {
	Sprites       []*ebiten.Image
	X, Y          float64
	Speed         float64
	FrameIndex    int
	FrameDelay    time.Duration
	LastUpdate    time.Time
	Visible       bool
	IsStopped     bool // New field to indicate if the enemy is stopped
	StopDuration  time.Duration
	OriginalSpeed float64 // New field to store the original speed
	IsSlowed      bool

	Cooldown int
	Hit      bool
}

func NewEnemy(sprites []*ebiten.Image, x, y float64) *Enemy {
	return &Enemy{
		Sprites:       sprites,
		X:             x,
		Y:             y,
		Speed:         enemySpeed,
		FrameIndex:    0,
		FrameDelay:    100 * time.Millisecond,
		LastUpdate:    time.Now(),
		Visible:       true,
		OriginalSpeed: enemySpeed,
	}
}
func (e *Enemy) Hitt(g *Game) {
	if e.Visible {
		e.Visible = false
		g.playerMoney += 100 // Increase the player's money
	}
}

func (e *Enemy) Update() {
	now := time.Now()
	if e.IsStopped {
		e.StopDuration += now.Sub(e.LastUpdate)
		if e.StopDuration >= 5*time.Second {
			e.IsStopped = false
			e.StopDuration = 0
		}
	} else if e.Visible && e.X > 100 {
		e.X -= e.Speed
	} else {
		e.Visible = false
	}

	if e.Hit { // Check if the enemy has been hit
		e.Visible = false // Mark the enemy as not visible if it has been hit
	}

	e.LastUpdate = now
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	if e.Visible {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X, e.Y)
		screen.DrawImage(e.Sprites[e.FrameIndex], op)
	}

}

func LoadEnemySprites() ([]*ebiten.Image, error) {
	sheet, _, err := ebitenutil.NewImageFromFile(enemySpriteSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to load enemy sprite sheet: %w", err)
	}

	var sprites []*ebiten.Image
	// Only load sprites from the third row (index 2)
	i := 2
	for j := 0; j < enemySpriteColumns; j++ {
		x := j * enemySpriteWidth
		y := i * enemySpriteHeight
		sprite := sheet.SubImage(image.Rect(x, y, x+enemySpriteWidth, y+enemySpriteHeight)).(*ebiten.Image)
		sprites = append(sprites, sprite)
	}

	return sprites, nil
}

type Shot struct {
	Image   *ebiten.Image
	X, Y    float64
	Speed   float64
	Visible bool
}

type Game struct {
	level          *tiled.Map
	tileDict       map[uint32]*ebiten.Image
	builtTowers    map[[2]int]int
	towerImages    []*ebiten.Image
	selectedTower  int
	towerButtons   []Button
	playerMoney    int
	playerHealth   int
	enemies        []*Enemy
	lastEnemySpawn time.Time
	shotImage      *ebiten.Image
	shots          []*Shot
	lastTowerShoot map[int]time.Time // Add this field to track tower shoot times
	playerScore    int
}

func NewGame() (*Game, error) {
	tmxMap, err := tiled.LoadFile(mapPath)
	lastTowerShoot := make(map[int]time.Time)
	if err != nil {
		return nil, fmt.Errorf("error loading TMX map: %w", err)
	}

	tileDict := make(map[uint32]*ebiten.Image)
	for _, tileset := range tmxMap.Tilesets {
		for _, tile := range tileset.Tiles {
			img, _, err := ebitenutil.NewImageFromFile(tile.Image.Source)
			if err != nil {
				return nil, fmt.Errorf("failed to load tile image from source %s: %w", tile.Image.Source, err)
			}
			tileDict[tile.ID+tileset.FirstGID] = img
		}
	}

	towerImages := make([]*ebiten.Image, len(towerPaths))
	for i, path := range towerPaths {
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load tower image: %w", err)
		}
		towerImages[i] = img
	}

	towerButtons := make([]Button, len(towerPaths))
	for i, img := range towerImages {
		scaledImg := ebiten.NewImage(int(float64(img.Bounds().Dx())*buttonScale), int(float64(img.Bounds().Dy())*buttonScale))
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(buttonScale, buttonScale)
		scaledImg.DrawImage(img, op)

		buttonImage := ebiten.NewImage(selectionAreaWidth, buttonSize)
		xOffset := (selectionAreaWidth - scaledImg.Bounds().Dx()) / 2
		yOffset := (buttonSize - scaledImg.Bounds().Dy()) / 2
		op.GeoM.Reset()
		op.GeoM.Translate(float64(xOffset), float64(yOffset))
		buttonImage.DrawImage(scaledImg, op)

		towerButtons[i] = Button{
			Image: buttonImage,
			X:     0,
			Y:     i*(buttonSize+buttonMargin) + buttonMargin,
		}
	}

	return &Game{
		level:          tmxMap,
		tileDict:       tileDict,
		builtTowers:    make(map[[2]int]int),
		towerImages:    towerImages,
		selectedTower:  -1,
		towerButtons:   towerButtons,
		playerMoney:    2400,
		playerHealth:   100,
		lastEnemySpawn: time.Now(),
		enemies:        []*Enemy{},
		shots:          []*Shot{},
		lastTowerShoot: lastTowerShoot,
		playerScore:    0,
	}, nil
}

func NewShotFromTower3(g *Game, towerPos [2]int) *Shot {
	// Calculate the tower's position on the game map
	towerX := float64(towerPos[0]*g.level.TileWidth - 50)
	towerY := float64(towerPos[1]*g.level.TileHeight - 350)

	// Load the shot image for Tower3
	shotImage, _, err := ebitenutil.NewImageFromFile(shot2ImageFilePath)
	if err != nil {
		log.Fatalf("Failed to load shot2 image: %v", err)
	}

	// Create the shot instance at the tower's position
	newShot := &Shot{
		Image:   shotImage,
		X:       towerX,    // Set shot's X position to the tower's X position
		Y:       towerY,    // Set shot's Y position to the tower's Y position
		Speed:   shotSpeed, // Adjust the speed as needed
		Visible: true,
	}

	// Check for collision with enemies
	for _, enemy := range g.enemies {
		if enemy.Visible && isColliding(newShot, enemy) {
			enemy.Hit = true        // Mark the enemy as hit
			newShot.Visible = false // Make the shot invisible after hitting the enemy
			g.playerScore++
			break // Exit the loop after the first collision
		}
	}

	return newShot
}

func (g *Game) Update() error {
	// Handle mouse input for tower selection and placement
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// Check if a tower button is clicked
		for i, btn := range g.towerButtons {
			if x >= btn.X && x <= btn.X+selectionAreaWidth && y >= btn.Y && y <= btn.Y+buttonSize {
				g.selectedTower = i
				return nil
			}
		}

		// Check if a tower can be placed
		if g.selectedTower != -1 {
			tileX, tileY := (x-selectionAreaWidth)/g.level.TileWidth, y/g.level.TileHeight
			tilePos := [2]int{tileX, tileY}

			// Check if the target tile is not equal to TileIDNotAllowed
			if _, exists := g.builtTowers[tilePos]; !exists && tileX >= 0 && g.level.Layers[0].Tiles[tileY*g.level.Width+tileX].ID != 3 {
				towerIdx := g.selectedTower
				towerCost := towerCosts[towerIdx] // Get the cost based on the selected tower index

				if g.playerMoney >= towerCost {
					g.playerMoney -= towerCost
					g.builtTowers[tilePos] = towerIdx // Place the tower
				}
			}
			g.selectedTower = -1 // Deselect the tower after placement
		}
	}
	for _, shot := range g.shots {
		if shot.Visible {
			for _, enemy := range g.enemies {
				if enemy.Visible && isColliding(shot, enemy) {
					enemy.Hitt(g)        // Enemy is hit by a shot
					shot.Visible = false // Make the shot invisible after hitting the enemy
					g.playerScore++
				}
			}
		}
	}

	if (g.playerScore) < 10 {
		// Load the original map
		tmxMap, err := tiled.LoadFile(mapPath)
		if err != nil {
			return fmt.Errorf("error loading original TMX map: %w", err)
		}

		// Update the game's level and tileDict
		g.level = tmxMap

		// Update the currentMap variable to the original map path
		currentMap = mapPath
	} else {
		tmxMap, err := tiled.LoadFile(nextLevelMapPath)

		if err != nil {
			return fmt.Errorf("error loading original TMX map: %w", err)
		}

		// Update the game's level and tileDict
		g.level = tmxMap
		// Update the currentMap variable to the original map path
		currentMap = nextLevelMapPath
	}

	// Spawn a new enemy every 2 seconds
	if time.Since(g.lastEnemySpawn) >= 2*time.Second {
		enemySprites, _ := LoadEnemySprites()
		randomY := rand.Float64() * float64(g.level.Height*g.level.TileHeight-enemySpriteHeight)
		newEnemy := NewEnemy(enemySprites, float64(g.level.Width*g.level.TileWidth), randomY)
		g.enemies = append(g.enemies, newEnemy)
		g.lastEnemySpawn = time.Now()
	}

	// Update enemies and check for collisions with tiles 5 and 6
	activeEnemies := []*Enemy{}
	enemyReachedLeftEnd := false // Flag to track if an enemy reached the left end
	enemyReachedBase := false
	for _, enemy := range g.enemies {
		// Decrease cooldown if it's active
		if enemy.Cooldown > 0 {
			enemy.Cooldown--
		}

		// Enemy interaction with tower2
		if !enemy.IsStopped {
			for pos, idx := range g.builtTowers {
				if idx == 1 { //  index 1 corresponds to tower2
					towerX := float64(pos[0]*g.level.TileWidth + selectionAreaWidth)
					towerY := float64(pos[1] * g.level.TileHeight)

					// Check if enemy is within range of tower2
					if enemy.X >= towerX && enemy.X <= towerX+float64(g.level.TileWidth) &&
						enemy.Y >= towerY && enemy.Y <= towerY+float64(g.level.TileHeight) {
						// Stop enemy only if it's not in cooldown and not already stopped
						if enemy.Cooldown == 0 && enemy.StopDuration == 0 {
							enemy.IsStopped = true
							enemy.StopDuration = 1 // Stop duration set to 5 seconds
						}
						break
					}
				}
			}
		}

		// Handle decrementing stop duration and resetting IsStopped
		if enemy.IsStopped {
			if enemy.StopDuration > 0 {
				enemy.StopDuration -= 1 // Decrement the stop duration
			} else {
				enemy.IsStopped = false
				enemy.Cooldown = 10 // Set cooldown period after being stopped

			}
		}

	}

	for _, enemy := range g.enemies {
		enemyNearTower4 := false
		for pos, idx := range g.builtTowers {
			if idx == 3 { // Assuming index 3 corresponds to tower4
				towerX := float64(pos[0]*g.level.TileWidth + selectionAreaWidth)
				towerY := float64(pos[1] * g.level.TileHeight)
				// Check if enemy collides with tower4
				if enemy.X >= towerX && enemy.X <= towerX+float64(g.level.TileWidth) &&
					enemy.Y >= towerY && enemy.Y <= towerY+float64(g.level.TileHeight) {
					if !enemy.IsSlowed {
						enemy.Speed /= 5 // Slow down the enemy
						enemy.IsSlowed = true
					}
					enemyNearTower4 = true
					break
				}
			}
		}
		if !enemyNearTower4 && enemy.IsSlowed {
			// Restore the enemy's speed when it's no longer near tower4
			enemy.Speed = enemy.OriginalSpeed
			enemy.IsSlowed = false
		}
	}

	for _, enemy := range g.enemies {
		enemy.Update()

		if enemy.X-550 == 0 {
			enemyReachedBase = true
		}
		if enemy.X-100 <= 0 {
			enemyReachedLeftEnd = true
		} else {
			tileX := int(enemy.X) / g.level.TileWidth
			tileY := int(enemy.Y) / g.level.TileHeight

			if tileX < 0 || tileX >= g.level.Width || tileY < 0 || tileY >= g.level.Height {
				continue // Enemy is outside the map bounds
			}

			tileID := g.level.Layers[0].Tiles[tileY*g.level.Width+tileX].ID

			// If the enemy collides with tile 5 or 6, it will not be added to activeEnemies
			if tileID != 5 && tileID != 6 {
				activeEnemies = append(activeEnemies, enemy)
			}
		}
	}

	// Update the enemies slice
	g.enemies = activeEnemies

	// Decrease player health only once if an enemy reached the left end
	if enemyReachedBase {
		g.playerHealth = g.playerHealth - 1
	}
	if enemyReachedLeftEnd {
		g.playerHealth = g.playerHealth
	}

	// Update shots
	activeShots := []*Shot{}
	for _, shot := range g.shots {
		shot.X += shot.Speed
		if shot.X < float64(g.level.Width*g.level.TileWidth) {
			activeShots = append(activeShots, shot)
		}
	}
	g.shots = activeShots

	// Check if a tower can shoot
	for pos, idx := range g.builtTowers {
		if idx == 0 { // Assuming index 0 corresponds to Tower1
			// Check if it's time to shoot
			if lastShoot, ok := g.lastTowerShoot[idx]; !ok || time.Since(lastShoot) >= towerShootInterval {
				// Create a new shot from Tower1
				newShot := NewShotFromTower1(g, pos)
				g.shots = append(g.shots, newShot)

				// Update the last shoot time for Tower1
				g.lastTowerShoot[idx] = time.Now()
			}
		}
	}

	// Check for collisions between shots from Tower1 and enemies
	activeEnemies = []*Enemy{}
	for _, enemy := range g.enemies {
		if enemy.Visible {
			for _, shot := range g.shots {
				if shot.Visible && isColliding(shot, enemy) {
					enemy.Hit = true     // Mark the enemy as hit
					shot.Visible = false // Make the shot invisible after hitting the enemy
					g.playerScore++
				}
			}
		}

		if !enemy.Hit { // Add only enemies that haven't been hit back to activeEnemies
			activeEnemies = append(activeEnemies, enemy)
		}
	}

	// Update the enemies slice
	g.enemies = activeEnemies

	for pos, idx := range g.builtTowers {
		if idx == 2 { // Assuming index 2 corresponds to tower3
			// Check if it's time to shoot for tower3
			if lastShoot, ok := g.lastTowerShoot[idx]; !ok || time.Since(lastShoot) >= towerShootInterval {
				// Create a new shot from tower3
				newShot := NewShotFromTower3(g, pos)
				g.shots = append(g.shots, newShot)

				// Update the last shoot time for this tower
				g.lastTowerShoot[idx] = time.Now()
			}
		}
	}

	return nil
}
func isColliding(shot *Shot, enemy *Enemy) bool {
	// Calculate the boundaries of the shot and enemy
	shotLeft := shot.X
	shotRight := shot.X + float64(shot.Image.Bounds().Dx()-400)
	shotTop := shot.Y + 250
	shotBottom := shot.Y + float64(shot.Image.Bounds().Dy()-350)

	enemyLeft := enemy.X
	enemyRight := enemy.X + float64(enemySpriteWidth)
	enemyTop := enemy.Y
	enemyBottom := enemy.Y + float64(enemySpriteHeight-100)

	// Check for collision by comparing boundaries
	collision := shotRight > enemyLeft && shotLeft < enemyRight &&
		shotBottom > enemyTop && shotTop < enemyBottom

	// If there's a collision, update the shot's visibility
	if collision {
		shot.Visible = false
	}

	return collision
}

func NewShotFromTower1(g *Game, towerPos [2]int) *Shot {
	// Calculate the tower's position on the game map
	towerX := float64(towerPos[0]*g.level.TileWidth - 50)
	towerY := float64(towerPos[1]*g.level.TileHeight - 350)

	// Load the shot image
	shotImage, _, err := ebitenutil.NewImageFromFile(shotImageFilePath)
	if err != nil {
		log.Fatalf("Failed to load shot image: %v", err)
	}

	// Create the shot instance
	return &Shot{
		Image:   shotImage,
		X:       towerX,    // Set the shot's starting X position to the tower's X position
		Y:       towerY,    // Set the shot's starting Y position to the tower's Y position
		Speed:   shotSpeed, // Use the predefined shot speed
		Visible: true,
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	selectionAreaBg := ebiten.NewImage(selectionAreaWidth, g.level.Height*g.level.TileHeight)
	selectionAreaBg.Fill(color.RGBA{0x80, 0x80, 0x80, 0xff})
	screen.DrawImage(selectionAreaBg, nil)

	for _, layer := range g.level.Layers {
		for y := 0; y < g.level.Height; y++ {
			for x := 0; x < g.level.Width; x++ {
				tile := layer.Tiles[y*g.level.Width+x]
				if tile == nil || g.tileDict[tile.ID] == nil {
					continue
				}
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x*g.level.TileWidth+selectionAreaWidth), float64(y*g.level.TileHeight))
				screen.DrawImage(g.tileDict[tile.ID], opts)
			}
		}
	}
	for _, enemy := range g.enemies {
		if enemy.Visible {
			enemy.Draw(screen)
		}
	}

	// Draw shots
	for _, shot := range g.shots {
		if shot.Visible {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(shot.X, shot.Y)
			screen.DrawImage(shot.Image, op)
		}
	}

	for pos, idx := range g.builtTowers {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(pos[0]*g.level.TileWidth+selectionAreaWidth), float64(pos[1]*g.level.TileHeight))
		screen.DrawImage(g.towerImages[idx], opts)
	}

	for _, btn := range g.towerButtons {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(btn.X), float64(btn.Y))
		screen.DrawImage(btn.Image, opts)
	}

	for _, enemy := range g.enemies {
		enemy.Draw(screen)
	}

	// Draw shots
	for _, shot := range g.shots {
		if shot.Visible {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(shot.X, shot.Y)
			screen.DrawImage(shot.Image, op)
		}
	}

	// Display Player Money
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Money: $%d", g.playerMoney), 10, buttonSize+len(g.towerButtons)*(buttonSize+buttonMargin)+buttonMargin)

	// Display Player Health
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Health: %d", g.playerHealth), 10, buttonSize+len(g.towerButtons)*(buttonSize+buttonMargin)+buttonMargin+20)

	// Display Player Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.playerScore), 10, buttonSize+len(g.towerButtons)*(buttonSize+buttonMargin)+buttonMargin+40)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.level.Width * g.level.TileWidth, g.level.Height * g.level.TileHeight
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	// Set the window size based on the level dimensions
	ebiten.SetWindowSize(game.level.Width*game.level.TileWidth+selectionAreaWidth, game.level.Height*game.level.TileHeight)
	ebiten.SetWindowTitle("Tower Defense Game")

	// Start the game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
