const AuthInput = (props) => {
    const {
        type,
        id,
        placeholder,
        autoComplete,
        minLength,
    } = props

    return (
            <input className="auth-input" type={type} id={id} placeholder={placeholder} autoComplete={autoComplete} required minLength={minLength} />
    )
}

export default AuthInput