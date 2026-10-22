import { useEffect } from 'react'
import { useNavigate } from 'react-router'

const useRedirectIfLoggedIn = () => {
    const navigate = useNavigate()
    useEffect(() => {
        const checkAlreadyLogged = async () => {
            try {
                const res = await fetch('/auth/me')
                if (res.ok) {
                    navigate('/', { replace: true })
                }
            } catch {
                console.log('сервер не ответил')
            }
        }
        checkAlreadyLogged()
    }, [])
}

export default useRedirectIfLoggedIn