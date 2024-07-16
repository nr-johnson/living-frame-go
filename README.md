# DIY Living Picture Frame
This was a project I did as a gift for my mother-in-law and as a way to combine several skills I was learning into a single project.

The different skills required to complete this project are:
- Woodworking
- Computer hardware tinkering
- Programing
  - Go (Main server)
  - Python (GPIO Pin Configuration)
  - Javascript (Front end management and input)
- Linux Shell (Set up code and initiate app and GPIO services)
- PCB board soldering (Managing LED, input buttons and cooling fans)

## Code
The main app was written in Go and uses to Photoprism API to sync with an album within a Photoprism instance hosted on a local network. The app manages the wifi connection, authentication with Photoprism, and ensures the images match those within the album every few minutes.

The front end runs in a browser so all the code that manages the display images and user input is Javascript. There is a delay and fade duration setting, the ability to scroll back and forth through the images, sync the images, and shut off the device.

Buttons are manages using Python. The PI's GPIO pins made detecting input pretty easy.

Managing user input was a challange. I ended up creating a JSON file that all the different componants look to. When a button is pressed, the Python GPIO script updates the appropriate setting within that file.
The front end javascript will periodically send a request to the server for updated setting. The Go app will then get the JSON file and send the content down and the Javascript will initiate any changes. I couldn't find a more seamless way to get the three languages in the contexts to communiate with eachother. But the system works well.

## Screen
I used a screen that I took from an old Lenovo laptop. The power and display is managed through a control board I purchased from Amazon.
![Screen Testing](https://raw.githubusercontent.com/nr-johnson/living-frame-go/master/static/images/Screen_Test.jpg)

## PCB Board
I created this PCB board to help manage the companants needed for the LED, buttons and cooling fans. This was my first attempt at soldering. Needless to say, I went through a few iterations.
![PCB Front](https://raw.githubusercontent.com/nr-johnson/living-frame-go/master/static/images/PCB_Front.jpg)
![PCB Back](https://raw.githubusercontent.com/nr-johnson/living-frame-go/master/static/images/PCB_Back.jpg)
