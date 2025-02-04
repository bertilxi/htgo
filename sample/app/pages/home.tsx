import { Counter } from "@/components/counter";
import { Layout } from "@/components/layout";

interface Props {
  Name: string;
  InitialNumber: number;
}

export default function Home(props: Props) {
  return (
    <Layout>
      <div className="bg-green-500">
        <Counter defaultNum={props.InitialNumber} />

        <a href="/about">About</a>
      </div>
    </Layout>
  );
}
