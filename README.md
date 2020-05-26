
Inspired by Google Keep

Backend Service for a task-keeping app

API: https://shigoto-project.nn.r.appspot.com/api/v1/register  
App: https://shigoto-app.netlify.app/  
Frontend: https://github.com/peej-s/shigoto-frontend

# Sample Calls:  
To create a new user  
`curl -XPOST "https://shigoto-project.nn.r.appspot.com/api/v1/register" -d '{"username": "myUser", "password": "myPassword"}'`

To login a user  
`curl -XPOST "https://shigoto-project.nn.r.appspot.com/api/v1/login" -d '{"username": "myUser", "password": "myPassword"}'`

Response for registering or logging in a user:
`{"token":"NEW_TOKEN","userid":"USER_ID","expiry":"24_HOURS_FROM_NOW"}`

## Calls with token
These calls will fail if the token provided does not match the current token assigned to the user

To get a user's tasks  
`curl "https://shigoto-project.nn.r.appspot.com/api/v1/USER_ID/tasks" -H "Authorization: Bearer NEW_TOKEN"`

To add a new task for a user
`curl -XPOST "https://shigoto-project.nn.r.appspot.com/api/v1/USER_ID/tasks" -H "Authorization: Bearer NEW_TOKEN" -d '{"priority": some_int, "task": "some_string"}'`

To delete a task for a user
`curl -XDELETE "https://shigoto-project.nn.r.appspot.com/api/v1/USER_ID/tasks/TASK_ID" -H "Authorization: Bearer NEW_TOKEN"`
