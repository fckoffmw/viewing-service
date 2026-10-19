import AuthNav from "./AuthNav"
import SakuraFall from "./SakuraFall"

const AuthLayout = ({title, id, children, footer, onSubmit, error}) => {
    return (
         <>
            <SakuraFall />
            <AuthNav />
            <div className="auth-form">
                <h1>{title}</h1>
                <form id={id} onSubmit={onSubmit}>
                    {children}
                </form>
                {error && <div className="auth-error">{error}</div>}
                {footer}
            </div>
         </>
    )
}

export default AuthLayout