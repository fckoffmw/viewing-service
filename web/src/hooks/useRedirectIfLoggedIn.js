import { useEffect } from 'react'

const useRedirectIfLoggedIn = () => {
    useEffect(() => {
        const checkAlreadyLogged = async () => {
            try {
                const res = await fetch('/auth/me')
                if (res.ok) {
                    window.location.replace('/rooms.html')
                }
            } catch {
                console.log('сервер не ответил')
            }
        }
        checkAlreadyLogged()
    }, [])
}

export default useRedirectIfLoggedIn