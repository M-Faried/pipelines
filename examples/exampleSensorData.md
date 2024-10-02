# Real Life Example (Sensor Data)

In the examples folder you will find a full and complex example on how to use pipelines to process data streams efficiently. The example is the code to solve the below problem statement.

### Problem Statement:

- We have a sensor that sends data every 50ms in the structure SensorData. The data which should be saved in the database is the average of valid readings received from the sensor over 1 second span and the data should be approximated.

- If no data is received the average should still be logged with the most recent average value. And if the average is never calculated before, during the init phase, averages should still be logged with zero in the db.

- During the initialization of the sensor, the sensor goes through a calibration phase during which the sensor sends invalid temprature and humidity values which should be filtered. The invalid data value of the temprature less than -100 or the humidity is less than 0.

- The error data in the sensor reading should be monitored and calculated according to the following:

  - If the average error of the latest 5 temprature readings exceeds value of "10", the sensor should be recalibrated.

  - If the average error of the latest 10 humidity readings exceeds value of "15", the sensor should be recalibrated.
