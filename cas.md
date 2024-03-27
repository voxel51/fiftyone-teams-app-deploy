# FiftyOne Teams: Pluggable Authentication

The Pluggable Auth feature (sometimes referred to as CAS or Central Authentication Service) in FiftyOne Teams version 1.6.0 introduces a self-contained authentication system, eliminating the need for external dependencies like Auth0. This update is particularly advantageous for setups requiring an air-gapped or internal network environment, allowing FiftyOne Teams to operate in "internal mode." Key steps for setting up include configuring environmental variables for authentication, selecting authentication providers, and mapping users appropriately.

For organizations upgrading from earlier versions, the transition involves creating a new database dedicated to user and directory data. This process includes defining necessary permissions and adjusting settings to align with the **internal mode**'s requirements. The internal mode offers benefits such as support for multiple organizations and eliminates the reliance on external authentication services (Auth0).

The migration guide also covers the integration of authentication providers using the CAS REST API and details the process for migrating user data from Auth0. With the addition of JavaScript hooks, FiftyOne Teams can synchronize with corporate directories, offering customizable authentication and authorization solutions. This flexibility allows FiftyOne Teams to handle complex authentication scenarios.

## Using the Super Admin UI

This is where you configure the deployment wide configuration of FiftyOne Teams. When logging into Fiftyone Teams itself as an admin you are in the context of an organization, and settings there only apply to that organization. On the other hand, the Super Admin UI allows you to administer all organizations, and global configuration such as Identity Providers, Session timeouts, and JS hooks.

To login to this application you must navigate to **$YOUR_FIFTYONE_TEAMS_URL/cas/configurations** and provide the **FIFTYONE_AUTH_SECRET**in the top right of the screen to login.

## Fiftyone Auth Mode

With pluggable authentication comes a new setting called **FIFTYONE_AUTH_MODE**. This setting allows running Fiftyone Teams in two different modes:

**Legacy Mode Overview**

In Legacy Mode, the system uses Auth0 for user authentication and authorization, supporting only a single organization structure. This mode requires an external connection to Auth0 and follows an eventually consistent model. The configuration for identity providers and the persistence of user data in this mode is handled through Auth0, which includes support for SAML.

**Introduction to Internal Mode**

Transitioning to Internal Mode eliminates the need for Auth0, thereby removing the dependency on external services. This mode is capable of supporting multiple organizations and does not require external connectivity, making it suitable for environments where security is paramount or internet access is limited or not allowed. Unlike Legacy Mode, Internal Mode operates on an immediate consistency basis, ensuring that changes are reflected across the system instantly. Directory data is immediately written to MongoDB, and organizations have the autonomy to manage their Identity Provider Configuration. However, it is important to note that SAML support is not available in Internal Mode.

**Migrating from Legacy to Internal Mode**

The migration from Legacy to Internal Mode begins with configuring an authentication provider through the CAS REST API. This is a crucial step as it lays the foundation for the new authentication system. Following this, all existing user data must be migrated from Auth0, utilizing the management SDK or Auth0 API to ensure a comprehensive transfer of information.

For each user, the migration involves several steps executed via the CAS REST API:

1. Creating a FiftyOne Teams user profile via `POST /cas/api/users`.
2. Assigning the user to the default organization through a membership entry.
3. Linking the user's account to the new authentication provider by creating an account reference via `POST /cas/api/accounts`.

The final step in the migration involves changing the "fiftyone_auth_mode" setting from legacy to internal. This change officially activates Internal Mode, completing the migration process.

This migration enhances the system's autonomy by eliminating dependencies on external authentication services and allowing for a more controlled and secure management of user data and authentication processes.

## Installing FiftyOne Teams v1.6.0-beta.2

    **_NOTE: FiftyOne Teams v1.6.0-beta.2 has only been tested with new installs using the <code>internal</code> authentication mode. Upgrading an environment configured before v1.6.0-beta.2 has neither been tested nor documented. </em></strong>All of these instructions assume you already have a MongoDB instance prepared and a connection string available.

### **Helm**

    **_NOTE: FiftyOne Teams v1.6.0-beta.2 will create a new database in the same MongoDB instance as the fiftyone database, using the same connection string by default.  If you would like to use a different MongoDB instance, or different credentials, please set a new key-value pair in <code>secret.fiftyone</code> and update <code>casSettings.env.CAS_MONGODB_URI_KEY</code> to reflect the new key name.</em></strong>

A v1.6.0 `values.yaml` override example file is provided [here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/release/v1.6.0/helm/values.yaml)

`secret.fiftyone` notes:

* Auth0 secrets are no longer required in the `secret.fiftyone` configuration for `internal` authentication mode.  You can safely remove the following secrets when running in `internal` authentication mode:
  * `apiClientId`
  * `apiClientSecret`
  * `auth0Domain`
  * `clientId`
  * `clientSecret`
  * `organizationId`
* A new `fiftyoneAuthSecret` key has been added to the `secret.fiftyone` configuration and is required for CAS operations in any mode:
  * This secret is a random string used to authenticate to the CAS service.
  * This can be any string you care to use generated by any mechanism you prefer.
  * This is used for inter-service authentication and for the SuperUser to authenticate at the CAS UI to configure the Central Authentication Service.

`casSettings` notes:

All available `teams-cas` deployment configuration options can be found as part of the helm chart documentation and on [GitHub](https://arc.net/l/quote/ewewlkct)

The only authentication mode supported by v1.6.0-beta.2 is `internal` authentication mode.  This is configured by adding `FIFTYONE_AUTH_MODE: internal`to the `values.yaml` override file.

**Installing FiftyOne Teams v1.6.0-beta.2 via Helm**

1. Edit the `values.yaml` file and/or create the `secret.fiftyone.name` kubernetes secret with the appropriate key-value pairs:
    1. `encryptionKey`
    2. `fiftyoneDatabaseName`
    3. `fiftyoneAuthSecret`
    4. `mongodbConnectionString`
2. Edit the `values.yaml` file and update `teamsAppSettings.dnsName` to reflect the DNS Name your end users will use to connect to FiftyOne Teams.
3. Edit the `values.yaml` file and update any other configuration parameters related to your installation.  Guidance for v1.6.0 can be found [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/helm/fiftyone-teams-app).
4. Add and/or update the Voxel51 Helm repository
5. Install FiftyOne Teams using the pre-release v1.6.0-beta.2 Helm chart

### **Docker Compose**

    **_NOTE: FiftyOne Teams v1.6.0-beta.2 will create a new database in the same MongoDB instance as the fiftyone database, using the same connection string by default.  If you would like to use a different MongoDB instance, or different credentials, please uncomment and set the <code>CAS_DATABASE_URI</code> variable in your local <code>.env</code> configuration file.</em></strong>

Docker Compose yaml files and an environment file template are provided [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/docker/internal-auth)

`env.template` notes:

* Auth0 secrets are no longer required in the `.env`configuration for `internal` authentication mode.  You can safely remove the following secrets when running in `internal` authentication mode:
  * `AUTH0_API_CLIENT_ID`
  * `AUTH0_API_CLIENT_SECRET`
  * `AUTH0_AUDIENCE`
  * `AUTH0_CLIENT_ID`
  * `AUTH0_CLIENT_SECRET`
  * `AUTH0_DOMAIN`
  * `AUTH0_ISSUER_BASE_URL`
  * `AUTH0_ORGANIZATION`
* A new `FIFTYONE_AUTH_SECRET` variable has been added to the `.env` configuration and is required for CAS operations in any mode:
  * This secret is a random string used to authenticate to the CAS service.
  * This can be any string you care to use generated by any mechanism you prefer.
  * This is used for inter-service authentication and for the SuperUser to authenticate at the CAS UI to configure the Central Authentication Service.
* Updates to the [example nginx configurations](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/docker#environment-proxies) have been made to route traffic to the`/cas` endpoint to the `teams-cas` service.  If you are using alternative configurations to route traffic to the FiftyOne Teams services you will want to make appropriate configuration changes.

The only authentication mode supported by v1.6.0-beta.2 is `internal` authentication mode.  The Docker Compose configurations for this mode are in the`internal-auth` directory in the `fiftyone-teams-app-deploy` GitHub repository located [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/docker/internal-auth).

**Installing FiftyOne Teams v1.6.0-beta.2 via Docker Compose**

1. Clone the `release/v1.6.0` branch of the `fiftyone-teams-app-deploy` GitHub repository
2. `cd` into the `internal-auth` directory
3. copy the `env.template` file to `.env` for automatic use
4. Update the `.env` file to supply the required variables
    1. `BASE_URL`
    2. `FIFTYONE_API_URI`
    3. `FIFTYONE_DATABASE_URI`
    4. `FIFTYONE_AUTH_SECRET`
    5. `FIFTYONE_ENCRYPTION_KEY`
5. Update the `.env` file to update any other variables related to your installation.  Guidance for v1.6.0 configurations can be found [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/docker).
6. Create and edit `compose.override.yaml` to include any additional configuration related to your installation. e.g.

        ```
        name: fiftyone-teams
        services:
          fiftyone-app:
            image: voxel51/fiftyone-app-torch:v1.6.0-beta.2
        ```

7. Choose the plugins mode for your deployment, and select the appropriate compose file.  e.g.
8. Login to Docker Hub using the `voxeldocker` credentials provided by Voxel51
9. Pull the images required for your deployment
10. Deploy your FiftyOne Teams Stack
11. Use the [example nginx configurations](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/release/v1.6.0/docker), or some other mechanism of your choosing, to configure SSL termination and route end users to FiftyOne Teams services

## Getting started with Internal Mode

This section describes how to get up and running with the auth features in v1.6.0, including “air gapped” support (also called “internal mode”), removing dependencies on Auth0, and sourcing users internally. These steps are only required to run FiftyOne teams in “internal mode” and can be skipped if using Auth0.

1. Login to the SuperUser UI to configure your Authentication Provider / Identity Provider
2. Click on the “Admins” tab.
3. Click “Add admin” in the bottom left.
4. Specify your name email address as it appears in the Identity Provider that you will be configuring and then click “Add”.
5. Click on the “Identity Providers” tab at the top of the screen and then click “Add provider”.
6. Fill out the “Add identity provider”
    1. You can also click “Switch to advanced editor” to provide the full configuration as a JSON object.
7. In the “Profile callback” field, ensure that the mapping matches what is expected for your Identity Provider.
8. In the “Sign-in button style” specify how you would like users to see this login option in the UI.
    2. “Logo” is an optional field which allows you to provide a URL to a logo.
    3. “Background” and “Text color” allows you to define colors that will appear in the UI.
    4. Note that the text will be populated by the “Identity Provider Name” field at the top of the form.
9. Navigate to `$YOUR_FIFTYONE_TEAMS/datasets`
10. You should see the login choice for your newly configured authentication provider.
11. Before you login, make sure you have set your admin user in step 4. Otherwise you will need to remove this user from the database and try again.
12. Click the login button and provide the credentials that match the user defined as an admin in step 4
13. Once logged in, click on the icon in the top right corner then click “Settings”.
14. Click “Users” on the left side.
15. You should see yourself listed as an admin.

## Upgrading from &lt; 1.6.0 to 1.6.0

* You must create a new database to store directory and user information
  * By default this will be a new database in the same mongodb instance that hosts your current teams database
  * The credentials in `secrets.fiftyone.mongodbConnectionString` must have permissions to create a new database, create collections in that database, and create/modify indexes in those collections.
* When upgrading you should first get everything up and running in legacy mode. Then at any point in the future you can migrate from legacy mode to internal mode.

## Syncing with 3rd Party Directories

Below is an example of how to use JavaScript hooks to sync FiftyOne teams with a corporate directory such as Open Directory, LDAP, or Active Directory via an intermediary REST API or Identity Provider. Note that the recommended setup is to do this via OAuth/OIDC claims, however the example below illustrates a more intricate integration.

This example specifically addresses a scenario in which additional actions are performed during the **signIn** trigger, demonstrating how hooks can extend beyond simple authentication to interact with external APIs and internal services for complex user management and group assignment tasks. Here's a breakdown of the example:

### **Context Object**

The context object provides information about the current operation, including parameters like the user's details, and services (services) that offer utility functions and access to directory operations.

### **External API Integration**

* **getGroups**: This function calls an external API to retrieve a list of groups to which the signing-in user should be added. It utilizes the services.util.http.get method for making the HTTP request, demonstrating how external services can be queried within the hook.
* **addUserToGroup**: For each group retrieved from the external API, this function checks if the group exists in the organization's directory. If a group does not exist, it is created, and then the user is added to it. This process involves querying and modifying the organization's group directory, illustrating the hook's capability to perform complex operations like dynamic group management based on external data.

### **Error Handling**

* The try-catch block around the external API call and group manipulation logic ensures that errors do not prevent the user from signing in but are properly logged

### **Summary**

This hook example demonstrates a pattern for extending authentication flows in CAS with custom logic. By integrating with an external API to fetch group information and manipulating the organization's group memberships accordingly, it showcases the flexibility and extensibility of hooks in supporting complex, real-world authentication and authorization scenarios.

**Example**

# REST API

You can view the REST API Documentation by logging into the Super Admin UI (see above) or by directly visiting **$YOUR_FIFTYONE_TEAMS_URL/cas/api-doc**

# Configuration

## Feature Flags & Environment Variables

<table>
  <tr>
   <td><strong>ENV VAR</strong>
   </td>
   <td><strong>Type</strong>
   </td>
   <td><strong>Default</strong>
   </td>
   <td><strong>Deployments</strong>
   </td>
   <td><strong>Description</strong>
   </td>
  </tr>
  <tr>
   <td>CAS_BASE_URL
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-api
   </td>
   <td>Set this to the URL pointing to /cas/api of teams-cas service. I.e. <code>teams-cas-k8-service-name/cas/api</code>
   </td>
  </tr>
  <tr>
   <td>CAS_DATABASE_NAME
   </td>
   <td><code>String</code>
   </td>
   <td><code>"cas"</code>
   </td>
   <td>teams-cas
   </td>
   <td>Which mongodb database to use when storing FiftyOne Teams user data.
   </td>
  </tr>
  <tr>
   <td>CAS_DEFAULT_USER_ROLE
   </td>
   <td><code>Enum</code>
   </td>
   <td><code>"GUEST"</code>
   </td>
   <td>teams-cas
   </td>
   <td>GUEST, COLLABORATOR, MEMBER, ADMIN
   </td>
  </tr>
  <tr>
   <td>CAS_MONGODB_URI
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>The connection string to your mongodb server.
   </td>
  </tr>
  <tr>
   <td>CAS_URL
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>Set this to the URL pointing to the entry point of FiftyOne Teams. I.e. <code><a href="https://fiftyone.acme.com/cas/api/auth">https://fiftyone.acme.com</a></code>
   </td>
  </tr>
  <tr>
   <td>DEBUG
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>Controls debug logging if CAS
<p>
Example value:
<p>
<code>DEBUG=cas:*</code> - shows all cas logs
<p>
<code>DEBUG=cas:*:info</code> - shows only cas info logs
<p>
<code>DEBUG=cas:*,-cas:*:debug</code> - shows all cas logs except debug logs
   </td>
  </tr>
  <tr>
   <td>FEATURE_FLAG_ENABLE_INVITATIONS
   </td>
   <td><code>Boolean</code>
   </td>
   <td><code>true</code>
   </td>
   <td>teams-app, teams-api
   </td>
   <td>When true admins may invite users by email to onboard into the system. NOTE: This is currently not supported for internal mode.
   </td>
  </tr>
  <tr>
   <td>FIFTYONE_AUTH_MODE
   </td>
   <td><code>Enum</code>
   </td>
   <td><code>legacy</code>
   </td>
   <td>teams-cas
   </td>
   <td>legacy or internal - as described above.
   </td>
  </tr>
  <tr>
   <td>FIFTYONE_AUTH_SECRET
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas, teams-app, teams-api, fiftyone-app,
<p>
teams-plugins
   </td>
   <td>Generate a random secret for this value. This will be used to authenticate all requests to the system. This value should be regularly rotated.
   </td>
  </tr>
  <tr>
   <td>NEXTAUTH_URL
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>Set this to the full url pointing to the NEXTAUTH entrypoint eg. “<code><a href="https://fiftyone.acme.com/cas/api/auth">https://fiftyone.acme.com/cas/api/auth</a></code>”
   </td>
  </tr>
  <tr>
   <td>GLOBAL_AGENT_HTTP_PROXY
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-cas, teams-app
   </td>
   <td>The optional HTTP URL for configuring a proxy.
   </td>
  </tr>
  <tr>
   <td>GLOBAL_AGENT_HTTPS_PROXY
   </td>
   <td><code>URL</code>
   </td>
   <td>
   </td>
   <td>teams-cas, teams-app
   </td>
   <td>The optional HTTPS URL for configuring a proxy.
   </td>
  </tr>
  <tr>
   <td>GLOBAL_AGENT_NO_PROXY
   </td>
   <td><code>String (CSV)</code>
   </td>
   <td>
   </td>
   <td>teams-cas, teams-app
   </td>
   <td>The environment variable <code>GLOBAL_AGENT_NO_PROXY</code> value should be a comma-separated list of Docker Compose services that may communicate without going through a proxy server. By default these service names are
<ul>

<li>fiftyone-app

<li>teams-api

<li>teams-app

<li>teams-plugins

<li>teams-cas
</li>
</ul>
   </td>
  </tr>
</table>

<table>
  <tr>
   <td colspan="5" ><em>Legacy Mode Configuration</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_AUTH_CLIENT_ID
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td><em>Same as AUTH0_CLIENT_ID from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_AUTH_CLIENT_SECRET
   </td>
   <td><code>String</code>
   </td>
   <td><code>null</code>
   </td>
   <td>teams-cas
   </td>
   <td><em>Same as AUTH0_CLIENT_SECRET from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_DOMAIN
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td><em>Existing env from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_ISSUER_BASE_URL
   </td>
   <td><code>String</code>
   </td>
   <td><code>null</code>
   </td>
   <td>teams-cas
   </td>
   <td><em>Existing env from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_MGMT_CLIENT_ID
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td><em>Same as AUTH0_API_CLIENT_ID from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_MGMT_CLIENT_SECRET
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td><em>Same as AUTH0_API_CLIENT_SECRET from the past</em>
   </td>
  </tr>
  <tr>
   <td>AUTH0_ORGANIZATION
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td><em>Existing env from the past</em>
   </td>
  </tr>
  <tr>
   <td>TEAMS_API_DATABASE_NAME
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>same as FIFTYONE_DATABASE_NAME without the autoconsumption by <code>fiftyone</code>
   </td>
  </tr>
  <tr>
   <td>TEAMS_API_MONGODB_URI
   </td>
   <td><code>String</code>
   </td>
   <td>
   </td>
   <td>teams-cas
   </td>
   <td>same as FIFTYONE_DATABASE_URI without the autoconsumption by <code>fiftyone</code>
   </td>
  </tr>
</table>

## Config CAS API / Superuser

<table>
  <tr>
   <td><strong>Setting</strong>
   </td>
   <td>Type
   </td>
   <td>Default
   </td>
   <td>Description
   </td>
  </tr>
  <tr>
   <td><code>authentication_providers</code>
   </td>
   <td><code>Array</code>
   </td>
   <td><code>[]</code>
   </td>
   <td>A list of definitions of OIDC and/or OAuth providers.
   </td>
  </tr>
  <tr>
   <td><code>authentication_provider.profile</code>
   </td>
   <td><code>String</code> (parseable to JS function)
   </td>
   <td><code>null</code>
   </td>
   <td>When provided this function is called to map the external user_info to the internal fiftyone user/account. NOTE: function calls, the while keyword and other JS-specific syntax is not allowed in these functions.
   </td>
  </tr>
  <tr>
   <td><code>session_ttl</code>
   </td>
   <td><code>Number</code>
   </td>
   <td><code>300</code>
   </td>
   <td>Time in seconds for sessions to live after which users will be forced to log out. Must be greater than 120 seconds to support refreshing of user session using refresh token.
   </td>
  </tr>
  <tr>
   <td><code>js_hook_enabled</code>
   </td>
   <td><code>Boolean</code>
   </td>
   <td><code>true</code>
   </td>
   <td>When set to False, configured JavaScript hooks will not be invoked.
   </td>
  </tr>
  <tr>
   <td><code>js_hook</code>
   </td>
   <td><code>String</code> (parseable to a single JS function)
   </td>
   <td><code>null</code>
   </td>
   <td>JavaScript hook which is invoked on several CAS events described in JS Hooks section below
   </td>
  </tr>
</table>

## JavaScript Hooks Documentation for CAS

This documentation outlines the JavaScript hook implementation for the Custom Authentication Service (CAS). As a CAS superuser, you are able to define JavaScript functions that integrate with various authentication flows within CAS, customizing the authentication processes.

**Overview**

JavaScript hooks in CAS allow superusers to programmatically influence authentication flows, including sign-in, sign-up, JWT handling and content, redirection, and session management. This document describes the available hooks, their triggers, expected return types, and contextual information provided to each hook.

**Example JavaScript Hook**

**Actionable Triggers**

<table>
  <tr>
   <td><strong>Trigger</strong>
   </td>
   <td><strong>Description</strong>
   </td>
   <td><strong>Return Type</strong>
   </td>
  </tr>
  <tr>
   <td><code>signIn</code>
   </td>
   <td>Invoked when a user signs in. If the hook returns false or error is thrown, sign-in will be prevented.
   </td>
   <td>Boolean
   </td>
  </tr>
  <tr>
   <td><code>signUp</code>
   </td>
   <td>Invoked when a new user signs in for the very first time. If the hook returns false or an error is thrown, sign-in will be prevented and a user/account will not be created.
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><code>jwt</code>
   </td>
   <td>Invoked when JWT is created (on signIn, signUp, refresh token). The returned object will override payload of default JWT payload. If an error is thrown, the session will be expired and user will be redirected to sign-in
   </td>
   <td>Object | undefined
   </td>
  </tr>
  <tr>
   <td><code>redirect</code>
   </td>
   <td>Invoked post signIn or signOut. The user will be redirected to the URL/Path returned from the hook.
   </td>
   <td>String (URL)
   </td>
  </tr>
  <tr>
   <td><code>session</code>
   </td>
   <td>Invoked when a request for a session (on signIn, signUp, refresh token) is received.
   </td>
   <td>
   </td>
  </tr>
</table>

**Event-Only Triggers**

<table>
  <tr>
   <td><strong>Trigger</strong>
   </td>
   <td><strong>Description</strong>
   </td>
  </tr>
  <tr>
   <td><code>signOut</code>
   </td>
   <td>Invoked when a user signs out.
   </td>
  </tr>
  <tr>
   <td><code>createUser</code>
   </td>
   <td>Invoked when the adapter is asked to create a user.
   </td>
  </tr>
  <tr>
   <td><code>linkAccount</code>
   </td>
   <td>Invoked when an account is linked to a user.
   </td>
  </tr>
</table>

**JavaScript Hooks Contextual Parameters**

<table>
  <tr>
   <td>
<h2><strong>Parameter</strong></h2>

   </td>
   <td>
<h2><strong>Description</strong></h2>

   </td>
   <td>
<h2><strong>Available in Triggers</strong></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>token</code></h2>

   </td>
   <td>
<h2>The payload of a JWT token.</h2>

   </td>
   <td>
<h2><code>signIn<strong>, </strong>signUp<strong>, </strong>jwt<strong>, </strong>session<strong>, </strong>signOut<strong>, </strong>linkAccount<strong>, </strong>createUser</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>user</code></h2>

   </td>
   <td>
<h2>The signed-in user object.</h2>

   </td>
   <td>
<h2><code>signIn<strong>, </strong>signUp<strong>, </strong>jwt<strong>, </strong>linkAccount</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>account</code></h2>

   </td>
   <td>
<h2>The account from an identity provider.</h2>

   </td>
   <td>
<h2><code>signIn<strong>, </strong>signUp<strong>, </strong>jwt<strong>, </strong>linkAccount</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>profile</code></h2>

   </td>
   <td>
<h2>The profile from an identity provider.</h2>

   </td>
   <td>
<h2><code>signIn<strong>, </strong>signUp<strong>, </strong>jwt<strong>, </strong>linkAccount</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>isNewUser</code></h2>

   </td>
   <td>
<h2>True if a user is signing in for the first time.</h2>

   </td>
   <td>
<h2><code>jwt</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>trigger</code></h2>

   </td>
   <td>
<h2>Specifies the current trigger event.</h2>

   </td>
   <td>
<h2><code>jwt</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>Session</code></h2>

   </td>
   <td>
<h2>The session object.</h2>

   </td>
   <td>
<h2><code>session</code></h2>

   </td>
  </tr>
  <tr>
   <td>
<h2><code>services</code></h2>

   </td>
   <td>
<h2>Provides access to various services.</h2>

   </td>
   <td><strong>all</strong>
   </td>
  </tr>
</table>

**Services**

<table>
  <tr>
   <td><strong>Service</strong>
   </td>
   <td><strong>Description</strong>
   </td>
  </tr>
  <tr>
   <td><code>services.util.http</code>
   </td>
   <td>Provides <code>get</code>, <code>post</code>, <code>put</code>, and <code>delete</code> functions for making <code>HTTP</code> requests from a JS Hook.
   </td>
  </tr>
  <tr>
   <td><code>services.userContext</code>
   </td>
   <td>Object containing information about the user performing the current action.
   </td>
  </tr>
  <tr>
   <td><code>services.directory</code>
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><code>services.directory.users</code>
   </td>
   <td>The <code>UserService</code> - providing methods for interacting with the directory of users.
   </td>
  </tr>
  <tr>
   <td><code>services.directory.groups</code>
   </td>
   <td>The <code>GroupsService </code>- providing methods for interacting with the directory of groups.
   </td>
  </tr>
  <tr>
   <td><code>services.config</code>
   </td>
   <td>The <code>ConfigService </code>- providing methods for reading and writing the <code>AuthenticationConfig</code>.
   </td>
  </tr>
  <tr>
   <td><code>services.util</code>
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td><code>services.directory.orgs</code>
   </td>
   <td>The <code>OrgsService </code>- providing methods for interacting with the directory of organizations.
   </td>
  </tr>
  <tr>
   <td><code>services.webhookService</code>
   </td>
   <td>Experimental
   </td>
  </tr>
  <tr>
   <td><code>process.env['MY_ENV_VAR']</code>
   </td>
   <td>Syntax for reading environment variables in a JS Hook.
   </td>
  </tr>
</table>
