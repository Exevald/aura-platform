# Application registry

Единый реестр продуктовых приложений Aura. HTTP и gRPC API возвращают стабильные
UUID, код, имя, frontend route, backend target и признак `defaultEnabled`.
Реестр не владеет пользовательскими доступами: Access использует его как каталог.

```sh
mise run //modules/appreg:test
mise run //modules/appreg:integration-tests
```

Интеграционный тест проверяет одинаковый каталог через HTTP и gRPC.
