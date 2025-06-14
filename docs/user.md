# User

The main purpose of the user module is to identify a given user along with their authorization level.

## Registration

## Authentication

In this example repository, we use `user_password` as a way to authenticate the user.

## Session Management

## Security Model

In this project, we are trying to implement a Role Based Acces Control(RBAC)/Access Control List(ACL) for our users. The RBAC is simply a group of `role` that has `permissions` inside it that can be assigned to a single or multiple `users`.

To learn more about RBAC, you can look into these articles:

1. [Tailscale](https://tailscale.com/blog/rbac-like-it-was-meant-to-be).

```
|-------------|           |-------|           |-------|
| Permissions |           | Roles |           | Users |
|-------------|           |-------|           |-------|
|    P_1 ------------------> R_1 --------------> U_1  |
|    P_2 ------------------> R_2 --------^    |   .   |
|    P_3 ----------^      |   .   |           |   .   |
|     .       |           |   .   |           |   .   |
|    P_10----------v      |   .   |           |   .   |
|    P_11------------------> R_10 -------------> U_10 |
|------------|            |-------|           |-------|
```

As we caan see above, there are relations between `permissions`, `roles`, and `users`.

1. A `permission` can be attached to one or more `roles`.
1. A `role` can be attached to one or more `users`.
1. Because of things above, a given `user` can have more than one `role` and more than one `permission`.

While it looked simple, RBAC can be complex if mappings between `users`, `roles` and `permissions` are ambigous and can lead to the [role explosion](https://permify.co/post/role-explosion/) problem. And to maintain simplicity of this project(as an example) we would like to adopt the idea from [Tailscale](https://tailscale.com/blog/rbac-like-it-was-meant-to-be) and arrange our RBAC rules accordingly.

The Tailscale's concept revolve around:

1. Object types

	An object types is a `tag` that can be anything. It can be a file, or any other object that being tagged.

1. Role

	Describe humans in the identity system. For example `Finance`, `Software Engineering`, etc.

1. Entitlement

	A something

### Table Structure

To accomodate RBAC model, we will create table structures that supports mapping of `permissions` to `roles` and `roles` to `users`.

**Users**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| user_id | bigint | No | Yes |
| user_uuid | uuid | No | No |
| security_roles | bigint[] | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Security Roles**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| role_id | bigint | No | Yes |
| role_uuid | uuid | No | Yes |
| role_name | varchar | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Security Permission Keys**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| permission_key | varchar(30) | No | Yes |
| permission_type | varchar(20) | No | Yes |
| permission_key_description | text | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

**Permission Value(ENUM)**

| Name | Description |
|------|-------------|
| Read | Permission To Read |
| Write | Permission To Write |
| Delete | Permission To Delete |

**Security Role Permission**
| Column Name | Data Type | Nullable | Primary Key |
|------------|-----------|----------|-------------|
| role_id | bigint | No | Yes |
| permission_key | varchar | No | Yes |
| permission_values | permission_value[] | No | No |
| permission_bits_value | int | No | No |
| created_at | timestamptz | No | No |
| updated_at | timestamptz | Yes | No |

The goal of the data structure is to ensure each role can be assigned to a permission_key and define
the permissions freely inside it.

```text
    |-------|
	| Users |
	|-------|
	    |
        v
    |-------|    |------|
	| Roles | -> | Keys |
	|-------|    |------|
```

### Cache Data Structure

As we mainly use Redis for our `cache` store, we need to define hows the permissions and roles are being stored inside of the cache store.

### Authorization

So, how does the RBAC model and the tables data structures plays their role in `authorization` process?

To understand this, please make sure you are not skipping the [Authentication](#authentication) part and read them properly. In the `authentication` part, we know that we are storing sessions as a unique identifier for a user to stay authenticated to the system.