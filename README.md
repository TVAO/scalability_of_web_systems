# Scalability of Web Systems

This repository contains code and written solutions for the Scalability of Websystems course at the IT University of Copenhagen, held in 2017.

The language of choice is Go(lang) to write scalable, distributed services.

We write a Go service to query European Sentinel-2 satellite image data.

Project 1 involves designing a Go service on GCP that the client can use to get satellite data for a specific location. This typically involves specifying a latitude and longitude but it also supports human-like addresses (GeoCoding API).

Project 2 enables the ability to query satellite images in a geographical area, specified by two latitude and longitude bands. It also shows how we can profile (benchmark) our Go web service in terms of response time, memory usage, and throughput. A worker pool is implemented with goroutines and channels to scale the service. This permits the service to fetch thousands of satelitte images in seconds.

Project 3 extends the satelite image Go service with a geometry package. This permits the service to fetch and parse Geofabrik polygon data to construct S2 region covers, which can be used to fetch all satellite images of a country. It also implements a custom retry mechanism if some image fetches fail. The service can count the total number of Russian satellite images in seconds. 

The practice exam shows how to design a chat program with multiple connected users. It uses goroutines and channels to connect multiple clients to a chat server.

The exam demonstrates theoretical knowledge about building scalable systems.  
