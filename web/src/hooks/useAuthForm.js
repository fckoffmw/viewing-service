import { useState } from 'react'
import { useNavigate } from 'react-router'

const useAuthForm = (url, errorText) => {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    const navigate = useNavigate()

    const handleUsernameChange = (e) => {
        setUsername(e.target.value)
    }

    const handlePasswordChange = (e) => {
        setPassword(e.target.value)
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        setError('')
        setIsLoading(true)
        try {
            const res = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username: username.trim(), password: password }),
            })
            const data = await res.json()

            if (res.ok) {
                navigate('/')
                return
            }

            setError(data.error || errorText)
        } catch {
            setError('Ошибка соединения')
        } finally {
            setIsLoading(false)
        }
    }
    return { username, password, error, isLoading, handlePasswordChange, handleUsernameChange, handleSubmit }
}

export default useAuthForm
