package examples

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	pip "github.com/m-faried/pipelines"
)

// ExampleSensor is an example of how to use the pipeline to code the sensor data requirements mentioned in the exampleSensorData.md.

const (
	TEMPRATURE_AVG_ERROR_VALUES_COUNT = 5
	TEMPRATURE_AVG_ERROR_THRESHOLD    = 10
	TEMPERATURE_VALID_VALUE_MIN       = -100
	TEMPERATURE_VALID_VALUE_MAX       = 100

	HUMIDITY_AVG_ERROR_VALUES_COUNT = 10
	HUMIDITY_AVG_ERROR_THRESHOLD    = 15
	HUMIDITY_VALID_VALUE_MIN        = 0
	HUMIDITY_VALID_VALUE_MAX        = 100

	SENSOR_DATA_AVG_INTERVAL    = 1 * time.Second
	SENSOR_DATA_AVG_BUFFER_SIZE = 25 //Since the data is received every 50ms, the buffer size should be 20 (+5 tolerance) to store the data received for 1 second.

	REPLICAS_FILTER_STEP       = 5
	REPLICAS_TEMP_MONITOR_STEP = 1
	REPLICAS_HUMI_MONITOR_STEP = 1
	REPLICAS_AVG_DATA_STEP     = 3
	REPLICAS_SAVE_DB_STEP      = 1

	PIPE_CHANNEL_SIZE           = 100
	SIMULATED_DATA_GEN_INTERVAL = 50 * time.Millisecond
)

type SensorData struct {
	TemperatureValue      float64
	TemperatureErrorValue float64
	HumidityValue         float64
	HumidityErrorValue    float64
}

// filter step
func isValidSensorData(data *SensorData) bool {
	isValid := data.TemperatureValue >= TEMPERATURE_VALID_VALUE_MIN && data.HumidityValue >= HUMIDITY_VALID_VALUE_MIN
	if !isValid {
		fmt.Println("Discarding Invalid:", data)
	}
	return isValid
}

// called with every input sensor data
func tempratureErrorMonitor(data []*SensorData) (*SensorData, pip.BufferFlags) {

	if len(data) < TEMPRATURE_AVG_ERROR_VALUES_COUNT {
		return nil, pip.BufferFlags{}
	}

	sum := 0.0
	for _, d := range data {
		sum += d.TemperatureErrorValue
	}
	avg := sum / float64(len(data))

	if avg > TEMPRATURE_AVG_ERROR_THRESHOLD {
		fmt.Println("----------------------------------------------------")
		fmt.Println("----recalibrating sensor due to temprature error----")
		fmt.Println("----------------------------------------------------")
		fmt.Println()
		return nil, pip.BufferFlags{
			// flusing the buffer to avoid sending the recalibration signal multiple times.
			FlushBuffer: true,
		}
	}

	return nil, pip.BufferFlags{}
}

// called with every input sensor data
func humidityErrorMonitor(data []*SensorData) (*SensorData, pip.BufferFlags) {

	if len(data) < HUMIDITY_AVG_ERROR_VALUES_COUNT {
		return nil, pip.BufferFlags{}
	}

	sum := 0.0
	for _, d := range data {
		sum += d.HumidityErrorValue
	}
	avg := sum / float64(len(data))

	if avg >= HUMIDITY_AVG_ERROR_THRESHOLD {
		fmt.Println("----------------------------------------------------")
		fmt.Println("-----recalibrating sensor due to humidity error-----")
		fmt.Println("----------------------------------------------------")
		fmt.Println()
		return nil, pip.BufferFlags{
			// flusing the buffer to avoid sending the recalibration signal multiple times.
			FlushBuffer: true,
		}
	}

	return nil, pip.BufferFlags{}
}

// called every second.
func sensorAvgDataCalculation(data []*SensorData) (*SensorData, pip.BufferFlags) {

	// If no data is received, zero average log.
	if len(data) == 0 {
		result := &SensorData{
			TemperatureValue:      0,
			TemperatureErrorValue: 0,
			HumidityValue:         0,
			HumidityErrorValue:    0,
		}

		return result, pip.BufferFlags{
			SendProcessOuput: true,
		}
	}

	tempratureSum := 0.0
	tempraturErrorSum := 0.0
	humiditySum := 0.0
	humidityErrorSum := 0.0

	for _, d := range data {
		tempratureSum += d.TemperatureValue
		tempraturErrorSum += d.TemperatureErrorValue
		humiditySum += d.HumidityValue
		humidityErrorSum += d.HumidityErrorValue
	}

	tempratureAvg := tempratureSum / float64(len(data))
	humidityAvg := humiditySum / float64(len(data))
	tempratureErrorAvg := tempraturErrorSum / float64(len(data))
	humidityErrorAvg := humidityErrorSum / float64(len(data))

	approximatedData := &SensorData{
		TemperatureValue:      math.Round(tempratureAvg),
		TemperatureErrorValue: math.Round(tempratureErrorAvg),
		HumidityValue:         math.Round(humidityAvg),
		HumidityErrorValue:    math.Round(humidityErrorAvg),
	}

	return approximatedData, pip.BufferFlags{
		SendProcessOuput: true,
		// Not flushing the data so that we can calculate the average when there is no new data.
	}
}

// final (result) step.
func saveInDB(data *SensorData) {
	values := fmt.Sprintf("Temperature: %.2f, Humidity: %.2f", data.TemperatureValue, data.HumidityValue)
	errors := fmt.Sprintf("TemperatureError: %.2f, HumidityError: %.2f\n\n", data.TemperatureErrorValue, data.HumidityErrorValue)
	fmt.Println("------Saving data to database------")
	fmt.Println(values)
	fmt.Println(errors)
}

func createSensorDataPipeline() pip.IPipeline[*SensorData] {
	builder := &pip.Builder[*SensorData]{}

	filter := builder.NewStep(pip.StepFilterConfig[*SensorData]{
		Label:        "filter",
		Replicas:     REPLICAS_FILTER_STEP,
		PassCriteria: isValidSensorData,
	})

	tempratureErrorMonitorStep := builder.NewStep(pip.StepBufferConfig[*SensorData]{
		Label:                 "tempratureErrorMonitor",
		Replicas:              REPLICAS_TEMP_MONITOR_STEP,
		PassThrough:           true, // Since it is used for monitor.
		BufferSize:            TEMPRATURE_AVG_ERROR_VALUES_COUNT,
		InputTriggeredProcess: tempratureErrorMonitor,
	})

	humidityErrorMonitorStep := builder.NewStep(pip.StepBufferConfig[*SensorData]{
		Label:                 "humidityErrorMonitor",
		Replicas:              REPLICAS_HUMI_MONITOR_STEP,
		PassThrough:           true, // Since it is used for monitor.
		BufferSize:            HUMIDITY_AVG_ERROR_VALUES_COUNT,
		InputTriggeredProcess: humidityErrorMonitor,
	})

	avgDataStep := builder.NewStep(pip.StepBufferConfig[*SensorData]{
		Label:                        "avgData",
		Replicas:                     REPLICAS_AVG_DATA_STEP,
		PassThrough:                  false,
		BufferSize:                   SENSOR_DATA_AVG_BUFFER_SIZE,
		TimeTriggeredProcess:         sensorAvgDataCalculation,
		TimeTriggeredProcessInterval: SENSOR_DATA_AVG_INTERVAL,
	})

	saveInDBStep := builder.NewStep(pip.StepTerminalConfig[*SensorData]{
		Label:    "saveInDB",
		Replicas: REPLICAS_SAVE_DB_STEP,
		Process:  saveInDB,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepInputChannelSize: PIPE_CHANNEL_SIZE,
		TrackTokensCount:            false,
	}
	pipeline := builder.NewPipeline(pConfig,
		filter,
		tempratureErrorMonitorStep,
		humidityErrorMonitorStep,
		avgDataStep,
		saveInDBStep,
	)

	pipeline.Init()

	return pipeline
}

// ExampleSensor is an example of how to use the pipeline to code the sensor data requirements mentioned in the exampleSensorData.md.
func ExampleSensor() {

	// Set up a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancelCtx := context.WithCancel(context.Background())

	fmt.Println("Running the Sensor Pipeline...")
	pipeline := createSensorDataPipeline()
	pipeline.Run(ctx)

	// Goroutine to cancel the context when an interrupt signal is received
	go func() {
		<-sigChan
		fmt.Println("Interrupt signal received, cancelling context...")

		// cancel the context
		cancelCtx()

		// terminating the pipeline and clearning resources
		pipeline.Terminate()

		fmt.Println("Pipeline Terminated!!!")
	}()

	// Simulating the sensor data generation every 50ms
	go func() {
		ticker := time.NewTicker(SIMULATED_DATA_GEN_INTERVAL)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				fmt.Println("Data Generation Stopped!!!")
				return
			case <-ticker.C:
				data := generateSensorData()
				pipeline.FeedOne(data)
			}
		}
	}()

	<-ctx.Done()
	// Wait for the goroutines to finish too.
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Sensor Data Example Done!!!")
}

func generateSensorData() *SensorData {
	return &SensorData{
		TemperatureValue:      rand.Float64()*(TEMPERATURE_VALID_VALUE_MAX-TEMPERATURE_VALID_VALUE_MIN) + TEMPERATURE_VALID_VALUE_MIN,
		TemperatureErrorValue: rand.Float64() * 5, // Example error value
		HumidityValue:         rand.Float64()*(HUMIDITY_VALID_VALUE_MAX-HUMIDITY_VALID_VALUE_MIN) + HUMIDITY_VALID_VALUE_MIN,
		HumidityErrorValue:    rand.Float64() * 20, // Example error value
	}
}
