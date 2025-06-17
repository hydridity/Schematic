# Path Schema Validator

**Path Schema Validator** (Schematic) aims to be flexible, schema-driven tool for validating structured paths like strings against configurable, rule-based schemas.

Current design enforces naming conventions for secret management systems.

Possible other usecases might include naming conventions for access control policies, and resource organization standards in CI/CD pipelines,and other automated workflows.

## Features

- **Declarative Schemas:** Define path structures using a simple, human-readable schema language.
- **Variable and Set Support:** Reference variables and predefined sets in your schemas.
- **Modifiers:** Apply transformations to variables (e.g., strip prefixes) before validation.
- **Wildcard Matching:** Support for single and multi-segment wildcards.
- **Extensible:** Easily add new constraint types or modifiers.

## Example Use Cases

- Enforcing naming conventions for secrets or resources in CI/CD pipelines.
- Validating access control paths in automation workflows.
- Ensuring compliance with organizational structure for resource identifiers.
- Validating configuration keys or file paths in infrastructure projects.

## What this tool is not about
- Replacing regex

## How It Works

1. **Define a Schema:**  
   Write a schema that describes the allowed structure of your paths, using variables, sets, and modifiers.

2. **Configure Variables and Sets:**  
   Provide values for variables and sets via environment variables or configuration files.

3. **Validate Input:**  
   Use the tool to check if a given input string matches the schema. If it matches, validation succeeds; otherwise, it fails with error.

## Example

**Schema:**
```
$some_variable.strip_prefix("helm-")/$[technologies]/+
```

**Context variables:**
- `some_variable` (variable): `group1/helm-project1` such as $CI_PROJECT_PATH in gitlab runner
- `technologies` (variable set): `["mssql", "kafka", "postgres"]` - defined in config or database

**Input:**  
`group1/project1/postgres/admin`

**Result:**  
Validation succeeds if the input matches the schema after applying variable values and modifiers.

## License

TODO