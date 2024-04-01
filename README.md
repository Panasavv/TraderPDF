# TraderPDF

go run main.go <startDate> <endDate> <tag>

startDate and endDate should include the date the payment was made. 
Best practice is to make startDate the date the payment was made and endDate the next day.

Example:
If a payment was made on 21 March 2024 and we want to find the total euros a user with tag "past" made, we use this:
go run main.go 2024-03-21 2024-03-22 past

