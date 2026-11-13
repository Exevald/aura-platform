type Props = { title: string; detail: string }

export function FeatureView({title, detail}: Props) {
    return <section>
        <div className="page-title">
            <div><span className="eyebrow">Aura</span><h1>{title}</h1><p>{detail}</p></div>
        </div>
    </section>
}
