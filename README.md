**To obtain a GitHub token**, you need to generate a personal access token in your GitHub account. Follow these steps to create a personal access token:

GitHub Settings:

Go to your GitHub account and navigate to "Settings."
In the left sidebar, click on "Developer settings."
Access Tokens:

Under "Access tokens," click on "Generate token."
Token Configuration:

Provide a name for your token.
Select the scopes (permissions) that your token needs. For generating a token to access your own repositories and events, you might need the repo and user scopes.
Generate Token:

Click the "Generate token" button.
Copy Token:

Copy the generated token immediately. GitHub will not show it again.



After that
setup the .env
with username and github token

than run by
```go run main.go```
