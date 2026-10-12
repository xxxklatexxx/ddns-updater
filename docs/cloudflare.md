# Cloudflare

## Configuration

### Example

```json
{
  "settings": [
    {
      "provider": "cloudflare",
      "zone_identifier": "some id",
      "domain": "domain.com",
      "ttl": 600,
      "token": "yourtoken",
      "ip_version": "ipv4",
      "ipv6_suffix": ""
    }
  ]
}
```

### Example with email and Global API Key

```json
{
  "settings": [
    {
      "provider": "cloudflare",
      "zone_identifier": "some id",
      "domain": "domain.com",
      "ttl": 600,
      "email": "you@example.com",
      "key": "your-global-api-key",
      "ip_version": "ipv4",
      "ipv6_suffix": ""
    }
  ]
}
```

### Compulsory parameters

- `"zone_identifier"` is the Zone ID of your site, from the domain overview page written as *Zone ID*
- `"domain"` is the domain to update. It can be `example.com` (root domain), `sub.example.com` (subdomain of `example.com`) or `*.example.com` for the wildcard.
See [this issue comment for context](https://github.com/qdm12/ddns-updater/issues/243#issuecomment-928313949). This is left as is for compatibility.
- `"ttl"` integer value for record TTL in seconds (specify 1 for automatic)
- API Token `"token"` (recommended), configured with DNS edit permissions for your zone
- Alternatively, Email `"email"` and Global API Key `"key"` (legacy authentication)

The `"user_service_key"` setting is retained for backward compatibility, but is not one of the authentication methods documented by Cloudflare's current DNS API. Migrate to an API Token when possible.

### Optional parameters

- `"proxied"` can be set to `true` to use the proxy services of Cloudflare
- `"ip_version"` can be `ipv4` (A records), or `ipv6` (AAAA records) or `ipv4 or ipv6` (update one of the two, depending on the public ip found). It defaults to `ipv4 or ipv6`.
- `"ipv6_suffix"` is the IPv6 interface identifier suffix to use. It can be for example `0:0:0:0:72ad:8fbb:a54e:bedd/64`. If left empty, it defaults to no suffix and the raw temporary IPv6 address of the machine is used in the record updating. You might want to set this to use your permanent IPv6 address instead of your temporary IPv6 address.

Special thanks to @Starttoaster for helping out with the [documentation](https://gist.github.com/Starttoaster/07d568c2a99ad7631dd776688c988326) and testing.
