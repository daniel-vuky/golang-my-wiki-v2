# Google Login Setup Steps

## 1. Create a Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click on "Select a project" at the top of the page
3. Click "New Project"
4. Enter a project name (e.g., "My Wiki")
5. Click "Create"

## 2. Enable the Google+ API
1. In the Google Cloud Console, go to "APIs & Services" > "Library"
2. Search for "Google+ API" or "Google People API"
3. Click on it and click "Enable"

## 3. Configure OAuth Consent Screen
1. Go to "APIs & Services" > "OAuth consent screen"
2. Select "External" user type
3. Fill in the required information:
   - App name
   - User support email
   - Developer contact information
4. Click "Save and Continue"
5. Under "Scopes", add:
   - `.../auth/userinfo.email`
   - `.../auth/userinfo.profile`
6. Click "Save and Continue"
7. Add test users if needed
8. Click "Save and Continue"

## 4. Create OAuth 2.0 Credentials
1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. Choose "Web application"
4. Set up the OAuth consent screen if prompted
5. Configure the OAuth client:
   - Name: "My Wiki Web Client"
   - Authorized JavaScript origins:
     ```
     http://localhost:8080
     ```
   - Authorized redirect URIs:
     ```
     http://localhost:8080/auth/callback
     ```
6. Click "Create"
7. Save the generated Client ID and Client Secret

## 5. Update Configuration
1. Create or update your configuration file (e.g., `config.yaml`) with the Google credentials:
   ```yaml
   auth:
     google:
       client_id: "your-client-id"
       client_secret: "your-client-secret"
   ```

## Important Notes
- Keep your Client ID and Client Secret secure
- Never commit these credentials to version control
- For production, update the authorized origins and redirect URIs with your actual domain
- Make sure to add all necessary test users in the OAuth consent screen
- The OAuth consent screen needs to be verified by Google if you plan to make your app public 