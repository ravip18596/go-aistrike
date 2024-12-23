# go-aistrike-task
**how to use**

 - `docker-compose up -d` on root folder run it will start docker
   container with all required dependencies. 
 - `docker-compose down` on root folder to shut down.
  - `localhost:8080` to access go container from localhost
  - To start the program, please hit the heartbeart URL - [http://localhost:8080/](http://localhost:3333/)

## URLs

1. Booking a Ticket -
   ```http request
   curl --location 'http://localhost:8080/tickets' \
   --header 'Content-Type: application/json' \
   --data '{
   "user_id": 2,
   "user_name": "Ravi"
   }'
   ```
   1. Writing to mysql table
   2. Writing to redis
   3. Generating events
   4. Consuming events
2. Fetching user bookings -
   ```http request
   curl --location 'http://localhost:8080/users/2/tickets'
   ```
3. Cancelling a Ticket - 
   ```http request
   Pass the booking id generated in the book ticket API response
   curl --location --request DELETE 'http://localhost:8080/tickets/1'
   ```
   1. Writing to mysql table
   2. Writing to redis
   3. Generating events
   4. Consuming events

