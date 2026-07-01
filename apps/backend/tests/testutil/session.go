package testutil

const SessionCookieAdmin = "tokenjoy_session_member=m-admin"

func SessionCookie(memberID string) string {
	return "tokenjoy_session_member=" + memberID
}

func PlatformSessionCookie(operatorID string) string {
	return "tokenjoy_platform_session=" + operatorID
}
