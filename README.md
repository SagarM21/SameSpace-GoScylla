<h1 align="center">Go-Scylla</h1>

### Prerequisite

- Docker must be installed in your system.
- Golang must be installed in your system.

### How to run the project?

- Use following commands in the terminal:
  - <b>Step 1: Pull the ScyllaDB Docker Image</b>
    - docker pull scylladb/scylla
  - <b>Step 2: Create a ScyllaDB Container</b>
    - docker run --name scylla-local -p 9042:9042 -d scylladb/scylla
  - <b>Step 3: Check ScyllaDB Status</b>
    - docker exec -it scylla-local nodetool status
  - <b>Step 4: Access the CQL Shell</b>
    - docker exec -it scylla-local cqlsh
- git clone `https://github.com/SagarM21/SameSpace-GoScylla.git`
- cd SameSpace
- To start the server:
  - `go run main.go` in new terminal
  - If any error occurs regarding keyspace then run: - Run this command in docker terminal: `CREATE KEYSPACE IF NOT EXISTS todo
WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };`
- You can use <b>Postman/ThunderClient</b> to test/run the apis:
  - `POST` http://localhost:8080/todos (Create new todos)
    - Set content-type: application/json while calling the api.
  - `DEL`http://localhost:8080/todos/userId/todoId (Delete a specific todo through userId and todoId)
  - `GET` http://localhost:8080/todos/userId (Get todo based on userId)
  - `GET`http://localhost:8080/todos/userId/todoId (Get particular todo using userId and todoId)
  - `PUT` http://localhost:8080/todos/userId/todoId (Update todo using userId and todoId)
  - `Get` http://localhost:8080/todoStatus/pending (Get todos based on status {pending, completed})
  - `Get`http://localhost:8080/todos?page=1&size=10&status=completed (Pagination endpoint api)

### Screenshots:

