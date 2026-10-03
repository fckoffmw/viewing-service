const AuthSwitch = ({href, question, text}) => {
    return (
        <p className="auth-switch">{question} <a href={href}>{text}</a></p>
    )
   
}

export default AuthSwitch