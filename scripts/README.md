# Database Scripts Directory

This directory contains SQL scripts that can be executed through the DevHive Backend API for administrative tasks such as:

- Database resets
- Data seeding
- Schema migrations
- Database maintenance

## Usage

Scripts can be executed via the `/api/v1/database/execute-script` endpoint:

```bash
POST /api/v1/database/execute-script
Content-Type: application/json
Authorization: Bearer <your-jwt-token>

{
  "fileName": "script_name.sql"
}
```

## Available Scripts

- `reset_schema.sql` - Resets the database schema and sample data
- `seed_data.sql` - Seeds the database with initial data
- `migrate_v1_to_v2.sql` - Example migration script

## Security Notes

- All database operations require authentication
- Scripts are executed with the same permissions as the application database user
- Only place trusted SQL scripts in this directory
- Scripts should be idempotent when possible

## Script Format

Scripts should:
- Use standard SQL syntax
- Include proper error handling
- Be idempotent (safe to run multiple times)
- Include comments explaining the purpose
- Use transactions for multi-statement operations
