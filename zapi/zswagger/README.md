# ZSwagger

[![Go Reference](https://pkg.go.dev/badge/github.com/go-zing/gozz-kit/zapi/zswagger.svg)](https://pkg.go.dev/github.com/go-zing/gozz-kit/zapi/zswagger)

Generate [Swagger 2.0](https://swagger.io/specification/v2/) OpenAPI Documentation from Go interfaces and structs.

ZSwagger is part of the [gozz-kit](https://github.com/go-zing/gozz-kit) ecosystem and works seamlessly with gozz code generation tools to automatically create comprehensive API documentation.

## Features

- 🚀 **Automatic Generation**: Parse Go interfaces and structs to generate OpenAPI specs
- 🔄 **Recursive Type Support**: Handles complex nested and recursive data structures safely
- 📝 **Documentation Extraction**: Extracts documentation from Go comments
- 🎯 **Flexible Parameter Binding**: Configurable mapping of struct fields to path, query, header, and body parameters
- 🏗️ **Type Safety**: Full type-aware schema generation with proper validation
- ⚡ **Performance Optimized**: Memoized parsing prevents redundant work and ensures fast generation

## Installation

```bash
go get github.com/go-zing/gozz-kit/zapi/zswagger
```

## Quick Start

### 1. Define Your API Interface

```go
package api

import "context"

//go:generate gozz run -p "doc" -p "api" ./

// +zz:api:./
// User management API
type UserService interface {
    // +zz:api:get|/users
    // Get all users
    ListUsers(ctx context.Context) ([]User, error)

    // +zz:api:get|/users/{id}
    // Get user by ID
    GetUser(ctx context.Context, id int64) (User, error)

    // +zz:api:post|/users
    // Create a new user
    CreateUser(ctx context.Context, req CreateUserRequest) (User, error)
}

type User struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}
```

### 2. Generate Documentation

```go
package main

import (
    "encoding/json"
    "os"

    "github.com/go-zing/gozz-kit/zapi/zswagger"
    "github.com/go-zing/gozz-kit/zdoc"
)

func main() {
    swagger := zswagger.Parse(
        api.UserService{},
        zswagger.WithDocFunc(zdoc.TypesDoc(api.ZZ_types_doc).TypeFieldDoc),
    )

    // Write to file
    data, _ := json.MarshalIndent(swagger, "", "  ")
    os.WriteFile("swagger.json", data, 0644)
}
```

## Configuration Options

### Parameter Binding

Control how struct fields are mapped to OpenAPI parameters:

```go
swagger := zswagger.Parse(
    api.UserService{},
    zswagger.WithBindings(map[string]zswagger.Binding{
        "GET": {
            Path:   "uri",    // Fields tagged with "uri" go to path parameters
            Query:  "form",   // Fields tagged with "form" go to query parameters
            Header: "",       // No header parameters
            Body:   false,    // Don't use request body
        },
        "POST": {
            Path:   "uri",
            Query:  "",
            Header: "",
            Body:   true,     // Use entire struct as request body
        },
    }),
)
```

### Custom HTTP Mapping

If your API definitions don't follow the default `METHOD|PATH` format:

```go
swagger := zswagger.Parse(
    api.UserService{},
    zswagger.WithHttpCast(func(api zapi.Api) zapi.HttpApi {
        // Custom logic to convert API to HTTP API
        parts := strings.Split(api.Resource, "|")
        return zapi.HttpApi{
            Api:    api,
            Method: parts[0],
            Path:   parts[1],
        }
    }),
)
```

### Documentation Function

Customize how documentation is extracted:

```go
swagger := zswagger.Parse(
    api.UserService{},
    zswagger.WithDocFunc(func(typ reflect.Type, fieldName string) string {
        // Custom documentation extraction logic
        return getCustomDoc(typ, fieldName)
    }),
)
```

## Supported Types

ZSwagger automatically generates schemas for:

- **Basic Types**: `string`, `int`, `int64`, `uint`, `float`, `bool`
- **Time Types**: `time.Time` → `string` with `date-time` format
- **Network Types**: `net.IP` → `string` with `ipv4` format, `url.URL` → `string` with `uri` format
- **Binary Data**: `[]byte` → `string` with `base64` format
- **JSON Data**: `json.RawMessage` → `object`
- **Complex Types**: structs, maps, slices, arrays
- **Pointers**: Properly handles optional fields
- **Embedded Structs**: Flattens embedded fields
- **Recursive Types**: Safely handles circular references

## Custom Type Registration

Register custom schema generators for specific types:

```go
// Register a custom type
zswagger.RegisterSchemaType(reflect.TypeOf(MyCustomType{}), func(schema *spec.Schema) {
    schema.Typed("string", "custom-format")
})
```

## Generated Output

The generated Swagger spec includes:

- **Paths**: API endpoints with methods, parameters, and responses
- **Definitions**: Schema definitions for all types
- **Tags**: Organized API groups
- **Info**: Basic API information

Example generated spec:

```json
{
  "swagger": "2.0",
  "info": {
    "title": "api.UserService",
    "version": "unknown",
    "description": "This file is generated by gozz-kit-zswagger"
  },
  "paths": {
    "/users": {
      "get": {
        "operationId": "api.UserService.ListUsers",
        "responses": {
          "200": {
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/api.User"
              }
            }
          }
        }
      },
      "post": {
        "operationId": "api.UserService.CreateUser",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/api.CreateUserRequest"
            }
          }
        ],
        "responses": {
          "200": {
            "schema": {
              "$ref": "#/definitions/api.User"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "api.User": {
      "type": "object",
      "properties": {
        "id": {"type": "integer", "format": "int64"},
        "name": {"type": "string"},
        "email": {"type": "string"}
      }
    }
  }
}
```

## API Reference

### Functions

#### `Parse(iterator zapi.Iterator, options ...func(*Option)) *spec.Swagger`

Parse an API iterator and generate a Swagger specification.

**Parameters:**
- `iterator`: API iterator (usually a struct implementing zapi.Iterator)
- `options`: Optional configuration functions

**Returns:** `*spec.Swagger` - The generated OpenAPI specification

#### `RegisterSchemaType(typ reflect.Type, fn func(*spec.Schema))`

Register a custom schema generator for a specific Go type.

**Parameters:**
- `typ`: The reflect.Type to register
- `fn`: Function that configures the schema

### Types

#### `Option`

Configuration options for parsing.

```go
type Option struct {
    HttpCast func(api zapi.Api) zapi.HttpApi  // Custom HTTP mapping
    Bindings map[string]Binding               // Parameter binding rules
    DocFunc  func(reflect.Type, string) string // Documentation extraction
}
```

#### `Binding`

Parameter binding configuration for HTTP methods.

```go
type Binding struct {
    Path   string  // Struct tag for path parameters (e.g., "uri")
    Query  string  // Struct tag for query parameters (e.g., "form")
    Header string  // Struct tag for header parameters
    Body   bool    // Whether to use the entire struct as request body
}
```

### Option Functions

#### `WithHttpCast(fn func(api zapi.Api) zapi.HttpApi) func(*Option)`

Set custom HTTP API mapping function.

#### `WithBindings(bindings map[string]Binding) func(*Option)`

Set parameter binding rules for different HTTP methods.

#### `WithDocFunc(fn func(reflect.Type, string) string) func(*Option)`

Set custom documentation extraction function.

## Integration with Gozz

ZSwagger works best with the [gozz](https://github.com/go-zing/gozz) code generation tool:

1. Annotate your interfaces with `+zz:api:` comments
2. Run `gozz run -p "api" -p "doc" ./` to generate API metadata
3. Use ZSwagger to generate OpenAPI documentation

## Examples

See the [example_test.go](example_test.go) and [example.json](example.json) files for a complete working example with complex types, recursive structures, and various parameter bindings.

## Contributing

Contributions are welcome! Please see the main [gozz-kit repository](https://github.com/go-zing/gozz-kit) for contribution guidelines.

## License

This project is licensed under the same license as gozz-kit.