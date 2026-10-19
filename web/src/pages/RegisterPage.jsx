import AuthInput from '../components/AuthInput'
import AuthSwitch from '../components/AuthSwitch'
import AuthButton from '../components/AuthButton'
import AuthLayout from '../components/AuthLayout'
import useRedirectIfLoggedIn from '../hooks/useRedirectIfLoggedIn'
import useAuthForm from '../hooks/useAuthForm'

const RegisterPage = () => {
    useRedirectIfLoggedIn()
    
    const {
        username,
        password,
        error,
        isLoading,
        handlePasswordChange,
        handleUsernameChange,
        handleSubmit,
    } = useAuthForm('/auth/register', 'Ошибка регистрации')
    const buttonText = isLoading ? 'Регистрация...' : 'Зарегистрироваться'

    return (
        <AuthLayout onSubmit={handleSubmit} error={error} title="Регистрация" id="register-form" footer={<AuthSwitch question="Есть аккаунт?" to="/login" text="Вход" />}>
            <AuthInput value={username} onChange={handleUsernameChange} type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username" />
            <AuthInput value={password} onChange={handlePasswordChange} type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="new-password" />
            <AuthButton text={buttonText} isLoading={isLoading} />
        </AuthLayout>
    )
}

export default RegisterPage