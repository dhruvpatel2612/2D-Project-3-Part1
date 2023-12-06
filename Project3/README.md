# Project3 by Siris Neupane and Dhruv Patel

Tower Defense Game
Welcome to the Tower Defense Game, a strategic game where you must defend your base from waves of enemies using various types of defense towers. Manage your resources wisely, build towers strategically, and protect your base at all costs.
In the Tower Defense Game, you are tasked with defending your base from relentless waves of enemies. Here are the key elements of the game:

Getting Started
Prerequisites
Before you can play the game, ensure you have the following prerequisites:

Go installed on your system. You can download and install Go from the official Go website.
Installation
Clone this repository to your local machine:
git clone <repository-url>
Change to the project directory:

cd tower-defense-game
Install the required dependencies:
To run the game, execute the following command:
go run main.go
This will start the Tower Defense game in a window.
Or,
Simply run it on Goland or other IDEs for Go.

Game Controls
Left Mouse Button: Select the tower and select the tile to place tower on the map.
Mouse Cursor: Hover over tower buttons to select a tower type for placement.

Player Health
You have a limited amount of health, which is displayed on the screen. If enemies reach your base, they will reduce your health. Your goal is to prevent your health from reaching zero.

Enemy Waves
Enemies will spawn at one side of the map and make their way toward your base. Your job is to stop them using strategically placed defense towers.

Defense Towers
You have access to four different types of defense towers, each with unique abilities, including different damage, attack speed, and special functionalities. Place these towers on the map to fend off enemy waves.

Currency
You start with a modest amount of currency, which you can use to build and upgrade towers. Defeating enemies earns you additional currency, allowing you to strengthen your defenses.

Map Layout
The game maps feature a combination of open terrain and wall tiles. Towers can be placed on specific squares, blocking enemy paths. However, there is always at least one route for enemies to reach your base.
Also, map changes to next level map after you score 10 points (you can change the points and even add new maps)

Game Mechanics
Player Health
Your health is displayed on the screen. If it reaches zero, the game is over. Protect your health by preventing enemies from reaching your base.

Enemy Waves
Enemies spawn at one side of the map and move toward your base. They follow a path based on the layout of the map, trying to find their way to your base while avoiding obstacles.

Defense Towers
You have four types of defense towers at your disposal, each with its own characteristics:

Tower Type 1: This tower does moderate damage with a medium attack speed.
Tower Type 2: Tower 2 stops the enemy while on range.
Tower Type 3: Tower 3 does high damage with a slower attack.
Tower Type 4: This tower slows down the enemy.
Place these towers strategically on the map to maximize their effectiveness in stopping enemy waves.

Currency
You start the game with an initial amount of currency. Defeating enemies earns you additional currency, which you can use to build and upgrade towers. Manage your currency wisely to build a formidable defense.

Map Layout
The game maps feature a combination of open terrain and wall tiles. You can place defense towers on specific squares, blocking enemy paths. However, there is always at least one route for enemies to reach your base. Plan your tower placement carefully to ensure maximum coverage.

Game Rules
Build and upgrade towers to defend against waves of enemies.
Prevent enemies from reaching your base by strategically placing towers.
Manage your currency to build and upgrade towers effectively.
Each tower type has unique characteristics and abilities.
Enemies follow a path based on the map layout, but they will attempt to find a way to your base.
