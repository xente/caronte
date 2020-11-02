package main

import (
	"Caronte/dashboard"
	scheduler "Caronte/helper"
	"Caronte/metrics_publisher"
	"Caronte/orchestrator/discovery"
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {

	ctx := context.Background()

	logLevel := flag.String("log.level", "INFO", "Define Log level {DEBUG or PROD}. Default value prod")
	enableDashboard := flag.Bool("dashboard", false, "Activate Dashboard")
	dashboardPort := flag.Int("dashboard.port", 80, "Dashboard port listener")
	schedulerDiscoveryTime := flag.Int("service.scheduler.discovery.time", 30, "Seconds to raise scale logic")
	sqsMetricPublisherQueuename := flag.String("sqs.metic.publisher.queue.name", "", "")
	sqsMetricPublisherQueueTime := flag.Int("sqs.metic.publisher.queue.time", 5, "")

	flag.Parse()

	if *logLevel == "DEBUG" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	zap.S().Info("Caronte init")

	//Init Scheduled Routines
	worker := scheduler.NewScheduler()

	//Init Caronte Service Discovery
	serviceDiscovery, err := discovery.NewDiscovery()
	if err == nil {
		serviceDiscovery.CaronteServiceDiscovery(ctx)
		worker.Add(ctx, serviceDiscovery.CaronteServiceDiscovery, time.Second*time.Duration(*schedulerDiscoveryTime))
	} else {
		zap.S().Error(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	//TODO load metrics dynamicaly
	go metrics_publisher.Init(2112)
	go metrics_publisher.SQSrecordMetrics(*sqsMetricPublisherQueuename, *sqsMetricPublisherQueueTime)

	if *enableDashboard {
		go dashboard.Dashboard(*dashboardPort)
	}

	<-quit
	worker.Stop()
}
