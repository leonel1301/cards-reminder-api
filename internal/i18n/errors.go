package i18n

import "fmt"

type ErrorKey string

const (
	ErrUnauthenticated                  ErrorKey = "unauthenticated"
	ErrMissingAuthorizationHeader       ErrorKey = "missing_authorization_header"
	ErrInvalidToken                     ErrorKey = "invalid_token"
	ErrFailedToResolveUser              ErrorKey = "failed_to_resolve_user"
	ErrInvalidCardID                    ErrorKey = "invalid_card_id"
	ErrInvalidOwnerID                   ErrorKey = "invalid_owner_id"
	ErrCardNotFound                     ErrorKey = "card_not_found"
	ErrOwnerNotFound                    ErrorKey = "owner_not_found"
	ErrDeviceTokenNotFound              ErrorKey = "device_token_not_found"
	ErrInternalServerError              ErrorKey = "internal_server_error"
	ErrFailedToListCards                ErrorKey = "failed_to_list_cards"
	ErrFailedToListOwners               ErrorKey = "failed_to_list_owners"
	ErrFailedToResolveTimezone          ErrorKey = "failed_to_resolve_timezone"
	ErrFailedToBuildDashboard           ErrorKey = "failed_to_build_dashboard"
	ErrOwnerHasAssignedCards            ErrorKey = "owner_has_assigned_cards"
	ErrNoDeviceTokensRegistered         ErrorKey = "no_device_tokens_registered"
	ErrFailedToSendNotification         ErrorKey = "failed_to_send_notification"
	ErrUserNotFound                     ErrorKey = "user_not_found"
	ErrFailedToDeleteAccount            ErrorKey = "failed_to_delete_account"
)

func Error(lang string, key ErrorKey) string {
	lang = NormalizeLanguage(lang)

	messages := map[ErrorKey]map[string]string{
		ErrUnauthenticated: {
			"en": "unauthenticated",
			"es": "no autenticado",
		},
		ErrMissingAuthorizationHeader: {
			"en": "missing or invalid authorization header",
			"es": "encabezado de autorización ausente o inválido",
		},
		ErrInvalidToken: {
			"en": "invalid or expired token",
			"es": "token inválido o expirado",
		},
		ErrFailedToResolveUser: {
			"en": "failed to resolve user",
			"es": "no se pudo resolver el usuario",
		},
		ErrInvalidCardID: {
			"en": "invalid card id",
			"es": "id de tarjeta inválido",
		},
		ErrInvalidOwnerID: {
			"en": "invalid owner id",
			"es": "id de titular inválido",
		},
		ErrCardNotFound: {
			"en": "card not found",
			"es": "tarjeta no encontrada",
		},
		ErrOwnerNotFound: {
			"en": "owner not found",
			"es": "titular no encontrado",
		},
		ErrDeviceTokenNotFound: {
			"en": "device token not found",
			"es": "token de dispositivo no encontrado",
		},
		ErrInternalServerError: {
			"en": "internal server error",
			"es": "error interno del servidor",
		},
		ErrFailedToListCards: {
			"en": "failed to list cards",
			"es": "no se pudieron listar las tarjetas",
		},
		ErrFailedToListOwners: {
			"en": "failed to list owners",
			"es": "no se pudieron listar los titulares",
		},
		ErrFailedToResolveTimezone: {
			"en": "failed to resolve timezone",
			"es": "no se pudo resolver la zona horaria",
		},
		ErrFailedToBuildDashboard: {
			"en": "failed to build dashboard",
			"es": "no se pudo generar el resumen",
		},
		ErrOwnerHasAssignedCards: {
			"en": "owner has assigned cards",
			"es": "el titular tiene tarjetas asignadas",
		},
		ErrNoDeviceTokensRegistered: {
			"en": "no device tokens registered",
			"es": "no hay tokens de dispositivo registrados",
		},
		ErrFailedToSendNotification: {
			"en": "failed to send notification",
			"es": "no se pudo enviar la notificación",
		},
		ErrUserNotFound: {
			"en": "user not found",
			"es": "usuario no encontrado",
		},
		ErrFailedToDeleteAccount: {
			"en": "failed to delete account",
			"es": "no se pudo eliminar la cuenta",
		},
	}

	if byLang, ok := messages[key]; ok {
		if message, ok := byLang[lang]; ok {
			return message
		}
	}

	return string(key)
}

func ValidationErrorMessage(lang string, field, message string) string {
	lang = NormalizeLanguage(lang)

	translated := validationMessage(lang, message)
	return fmt.Sprintf("%s %s", field, translated)
}

func validationMessage(lang, message string) string {
	messages := map[string]map[string]string{
		"is required": {
			"en": "is required",
			"es": "es obligatorio",
		},
		"must be exactly 4 digits": {
			"en": "must be exactly 4 digits",
			"es": "debe tener exactamente 4 dígitos",
		},
		"must be between 1 and 31": {
			"en": "must be between 1 and 31",
			"es": "debe estar entre 1 y 31",
		},
		"cannot be empty": {
			"en": "cannot be empty",
			"es": "no puede estar vacío",
		},
		"cannot delete self owner": {
			"en": "cannot delete self owner",
			"es": "no se puede eliminar tu perfil de titular",
		},
	}

	if byLang, ok := messages[message]; ok {
		if translated, ok := byLang[lang]; ok {
			return translated
		}
	}

	return message
}
