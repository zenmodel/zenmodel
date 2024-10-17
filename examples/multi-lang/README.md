# Multi-Language Support in ZenModel

This example demonstrates how to use ZenModel's multi-language support, allowing you to create a Brain that incorporates both Go and Python processors.

## Structure

The example consists of the following files:

- `main.go`: The main Go file that sets up and runs the multi-language Brain.
- `a/b/c/setname.py`: A Python file containing the `SetNameProcessor`.
- `d/e/f/add.py`: A Python file containing the `AddProcessor`.

## How to Use

1. Ensure you have both Go and Python installed on your system.

2. Install the required Python package:

   ```
   pip install zenmodel
   ```

3. Run the Go program:

   ```
   go run main.go
   ```

## Code Explanation

### Python Processors

1. `SetNameProcessor` (in `a/b/c/setname.py`):
   - Initializes with a `lastname`.
   - Reads the `name` from memory, appends the `lastname`, and writes it back to memory.

2. `AddProcessor` (in `d/e/f/add.py`):
   - Initializes with two integers `a` and `b`.
   - Reads `name` and `date` from memory, performs addition, and writes the result to memory.

### Go Main Program

The `main.go` file demonstrates how to:

1. Create a multi-language blueprint.
2. Add neurons with Python processors.
3. Set up links between neurons.
4. Build and run a multi-language Brain.

Key points:

- Use `zenmodel.NewMultiLangBlueprint()` to create a multi-language blueprint.
- Use `bp.AddNeuronWithPyProcessor()` to add neurons with Python processors.
- Use `brainlite.BuildMultiLangBrain()` to build the multi-language Brain.

## Output

The program will output the result of the processors' operations, demonstrating how data flows through the multi-language Brain.

This example showcases ZenModel's ability to seamlessly integrate Go and Python processors within a single Brain, allowing developers to leverage the strengths of both languages in their applications.
