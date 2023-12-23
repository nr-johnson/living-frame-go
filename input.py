#!/usr/bin/python

# from pyautogui import hotkey
import RPi.GPIO as GPIO
from time import sleep

GPIO.setmode(GPIO.BOARD)

GPIO.setup(24, GPIO.OUT)



try:
  while True:
      GPIO.output(24, GPIO.HIGH)

except KeyboardInterrupt:
      GPIO.output(24, GPIO.LOW)
      GPIO.cleanup()
      pass



# GPIO.setup(29, GPIO.IN, pull_up_down=GPIO.PUD_DOWN)
# GPIO.setup(31, GPIO.IN, pull_up_down=GPIO.PUD_DOWN)
# GPIO.setup(32, GPIO.IN, pull_up_down=GPIO.PUD_DOWN)
# GPIO.setup(33, GPIO.IN, pull_up_down=GPIO.PUD_DOWN)

# GPIO.setup(11, GPIO.OUT, initial=GPIO.LOW)
# GPIO.setup(13, GPIO.OUT, initial=GPIO.LOW)
# GPIO.setup(15, GPIO.OUT, initial=GPIO.LOW)

# currentColor = 0
# pressing = False

# try:
#   while True:
#       if not pressing:
#             if GPIO.input(29) == GPIO.HIGH:
#                   pressing = 29
                  
#                   print('Pressed 29')
            
#             elif GPIO.input(31) == GPIO.HIGH:
#                   pressing = 31
                  
#                   print('Pressed 31')

#             elif GPIO.input(32) == GPIO.HIGH:
#                   pressing = 32
                  
#                   print('Pressed 32')
            
#             elif GPIO.input(33) == GPIO.HIGH:
#                   pressing = 33
                  
#                   print('Pressed 33')
      
#       elif GPIO.input(29) == GPIO.LOW and GPIO.input(31) == GPIO.LOW and GPIO.input(32) == GPIO.LOW and GPIO.input(33) == GPIO.LOW:
            
#             pressing = False
      
#       sleep(.01)

# except KeyboardInterrupt:
#       GPIO.output(11, GPIO.LOW)
#       GPIO.output(13, GPIO.LOW)
#       GPIO.output(15, GPIO.LOW)
#       GPIO.cleanup()
#       pass