import { useEffect, useState } from "react"
import { useNavigate } from "react-router"

const useRequireAuth = () => {
    const [username, setUsername] = useState('')
    const [id, setID] = useState('')

    const navigate = useNavigate()

    useEffect(() => {
        const checkNotLogged = async () => {
            try {
                const res = await fetch('/auth/me')
                if (!res.ok) {
                    navigate('/login', { replace: true })
                    return
                } else {
                    const data = await res.json()
                    setUsername(data.username)
                    setID(data.id)
                }
            } catch {
                console.log('ошибка соединения')
            }
        }
        checkNotLogged()
    }, [])

    return {username, id}
}

export default useRequireAuth