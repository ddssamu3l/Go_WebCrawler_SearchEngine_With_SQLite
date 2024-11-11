# Search Engine in Go

This is a Web-Crawler + Search Engine project made in Go.

The program, when run, will crawl the university website of the University of San Francisco, storing all pages within the host into a SQLite database.

## How the Web Crawler works:
The web crawler starts with one seed url and access the **robots.txt** file associated with the hostname to crawl ethically. 

Then, the web crawler will access every URL it can within the host in BFS fashion, as long as the URLs are within the same hostname as the seed.

The crawler will extract all of the words and URLs, storing them in a relational SQLite database

## How the Search Engine works:
The search engine runs in **https://localhost:8080**

The search engine runs parallel to the Web Crawler, accessing the database as the Web Crawler builds the database with a **Go routine**.

A **Tf-Idf algorithm** is used to search the user-entered text and matches with the closest result within the database.

## In-Memory-Inverted-Index vs SQLite
This project can be run with both an **In-Memory-Index database** for fast builds or a **relational SQLite database** for permanent data storage.

The user can freely choose which type of database is used. 

**Inverted-Index-Database:** apply the **'-mode=memory"** Go flag when starting the program **(e.g "go run . -mode=memory")**

**SQLite Database:** apply the **"-mode=database"** Go flag when starting the program **(e.g "go run . -mode=database)**

## Unit Testing
Methods in this project are **unit tested**.

Each file in the project has its own '_test.go' file associated with it to test the project's functionality using **mock servers**.

## Running this project
To run this project, make sure you have **Golang** installed on your computer.

From the root directory of the project, run "go run .". To re-build databases, add the **"-makeNewDatabase"** Go flag to the command.

After that, go to https://localhost:8080 to access the search engine!


<img width="495" alt="截屏2024-11-10 下午4 52 46" src="https://github.com/user-attachments/assets/25eb6399-58b4-4866-9597-f1188a1e2327">
