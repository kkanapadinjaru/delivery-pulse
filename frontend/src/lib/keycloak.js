import Keycloak from 'keycloak-js'

let _kc = null

/**
 * Initialises Keycloak using the config returned by /api/config.
 * Redirects to the Keycloak login page if not authenticated.
 * @param {{ keycloakUrl: string, clientId: string }} config
 * @param {(token: string) => void} onTokenRefresh  called whenever the token is silently refreshed
 * @returns {Promise<Keycloak>}
 */
export async function initKeycloak({ keycloakUrl, clientId }, onTokenRefresh) {
  _kc = new Keycloak({ url: keycloakUrl, realm: 'delivery-pulse', clientId })

  _kc.onTokenExpired = () => {
    _kc.updateToken(30)
      .then(refreshed => { if (refreshed) onTokenRefresh(_kc.token) })
      .catch(() => _kc.login())
  }

  await _kc.init({
    onLoad: 'login-required',
    pkceMethod: 'S256',
    checkLoginIframe: false,
  })

  return _kc
}

export function logout() {
  _kc?.logout()
}
