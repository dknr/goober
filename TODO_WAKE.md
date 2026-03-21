# Wake Command Test Todo

Remaining work to complete wake command unit tests:

1. Fix output capture: The wake command uses fmt.Printf/Println directly, not captured by cmd.SetOut/SetErr.
   Options:
   - Refactor wake command to use cmd.Printf/cmd.Println (requires changing command to access *cobra.Command)
   - Use os.Stdout redirection in tests (more invasive but isolated to tests)

2. Once output capture is fixed, run tests to verify all pass.

3. Consider adding more edge cases:
   - Invalid hostname in request
   - Non-JSON responses
   - Network timeouts

