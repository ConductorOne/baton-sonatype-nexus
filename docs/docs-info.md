While developing the connector, please fill out this form. This information is needed to write docs and to help other users set up the connector.

## Connector capabilities

1. What resources does the connector sync?

- **Users**: All users configured in the Nexus repository manager, including their profile information (first name, last name, email, status, source)
- **Roles**: All roles defined in Nexus, including their descriptions and source information

2. Can the connector provision any resources? If so, which ones? 

No, this connector does not support provisioning. It is read-only and only syncs existing users, roles, and their assignments from Sonatype Nexus.

## Connector credentials 

1. What credentials or information are needed to set up the connector? (For example, API key, client ID and secret, domain, etc.)

- **Host URL**: The URL of the Nexus instance (e.g., `http://localhost:8081` or `https://nexus.company.com`)
- **Username**: A Nexus username with administrative access
- **Password**: The password for the specified username

2. For each item in the list above: 

   * How does a user create or look up that credential or info? Please include links to (non-gated) documentation, screenshots (of the UI or of gated docs), or a video of the process. 

   **Host URL**: 
   - This is the URL where your Nexus instance is running
   - For local development: typically `http://localhost:8081`
   - For production: the full URL of your Nexus server (e.g., `https://nexus.company.com`)
   - Can be found in your browser when accessing the Nexus web interface

   **Username and Password**:
   - These are the credentials of an existing Nexus user account
   - The default admin account is typically `admin` with the password set during installation
   - For custom users, they must be created through the Nexus web interface or API
   - Documentation: [Nexus User Management](https://help.sonatype.com/repomanager3/security/users)

   * Does the credential need any specific scopes or permissions? If so, list them here. 

   The user account needs the following permissions:
   - **Read access to users**: To list all users in the system
   - **Read access to roles**: To list all roles and their definitions
   - **Read access to user-role assignments**: To determine which roles are assigned to which users

   * If applicable: Is the list of scopes or permissions different to sync (read) versus provision (read-write)? If so, list the difference here. 

   Not applicable - this connector is read-only and does not support provisioning.

   * What level of access or permissions does the user need in order to create the credentials? (For example, must be a super administrator, must have access to the admin console, etc.)  

   The user account used for the connector must have:
   - **Administrator privileges** or equivalent role that grants access to user and role management
   - Access to the Nexus Security API endpoints
   - The `nx-admin` role is typically sufficient for full access  
