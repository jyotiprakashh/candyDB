# CandyDB

## Features

- Simple and intuitive API for CRUD operations (Create, Read, Update, Delete)
- Secure storage of data in AWS S3
- Scalable architecture to handle large datasets
- Built-in support for concurrency and synchronization
- Easy integration with existing Go applications
- RESTful API for interacting with the database remotely

## Installation

To use CandyDB in your Go projects, you can easily include it as a module:

```bash
go get github.com/your-username/CandyDB
```


Then, import the package in your code:
```bash
import (
    "github.com/your-username/CandyDB"
)
```

Getting Started
Set up AWS Credentials:
Ensure you have AWS credentials set up and configured.

Set Environment Variables:
Create a .env file with the following variables:
```bash
AWS_S3_REGION=your-region
AWS_S3_BUCKET=your-bucket
AWS_S3_PREFIX=your-prefix
```

Creating a Database Instance
To create a new instance of Candy DB, use the following code:
```bash
db, err := candyDB.New(nil)
if err != nil {
    fmt.Printf("Error: %s\n", err)
}
```

Writing Data
To write data to Candy DB, use the Write method:
```bash
err := db.Write("users", "Jyoti", User{
    ID:      uuid.New().String(),
    Name:    "Name",
    Age:     "18",
    Contact: "1234567890",
    Company: "ABC",
    Address: Address{City: "city", State: "state", Country: "country", Pincode: "751024"},
})
if err != nil {
    fmt.Printf("Error: %s\n", err)
}
```

Reading Data
To read data from Candy DB, use the Read method:
```bash
var user User
err := db.Read("users", "Jyoti", &user)
if err != nil {
    fmt.Printf("Error: %s\n", err)
}
fmt.Println(user)
```

Deleting Data
To delete data from Candy DB, use the Delete method:
```bash
Deleting Data
To delete data from Candy DB, use the Delete method:
```

Deleting All Data
To delete all data from a collection, use the DeleteAll method:
```bash
err := db.DeleteAll("users")
if err != nil {
    fmt.Printf("Error: %s\n", err)
}
```

HTTP Endpoint
Candy DB provides an HTTP endpoint for uploading data. You can use the following code to create an HTTP handler:
```bash
http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
    demoUser := User{
        ID:      uuid.New().String(),
        Name:    "Demo",
        Age:     "30",
        Contact: "0987654321",
        Company: "XYZ",
        Address: Address{City: "DemoCity", State: "DemoState", Country: "DemoCountry", Pincode: "000000"},
    }
    err := db.Write("users", demoUser.ID, demoUser)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("Error uploading user: %s", err)))
        return
    }
    w.Write([]byte("Uploaded"))
})
```

Run the Application:
```bash
go run main.go
```

# Conclusion

Candy DB is a lightweight, cloud-based NoSQL database that provides a simple and efficient way to store and retrieve data in Amazon S3. Its Go-based implementation ensures high performance and scalability, making it suitable for a wide range of applications.
