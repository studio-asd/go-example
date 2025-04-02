# Role Based Access Control (RBAC)

The RBAC service provides a set of APIs for managing roles and permissions.

## Data Structures Design Sketch

GET /v1/login
POST /v1/login
essential_api


session_metadata
roles:[id=1;name=standard_user]
permissions:[essential_api=r,w]
