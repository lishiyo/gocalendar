### Calendar script

Quick script in Go to fetch all your calendars and list the events in terminal. 

### Development

- Enable the Calendar api for your project and download the credentials.json file.
- run `go run main.go`

The first time you run the sample, it will prompt you to authorize access:
- Browse to the provided URL in your web browser.
- If you are not already logged into your Google account, you will be prompted to log in. If you are logged into multiple Google accounts, you will be asked to select one account to use for the authorization.
- Click the Accept button.
- Copy the code you're given, paste it into the command-line prompt, and press Enter.

### TODOs

Command-line options:
- number of days to list
- ascii timeline view?

Switch to terminal interface like https://github.com/jroimartin/gocui