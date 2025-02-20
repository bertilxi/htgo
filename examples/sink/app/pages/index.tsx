import { Counter } from "@/components/counter";
import { Layout } from "@/components/layout";

interface Props {
  time: string;
  route: string;
}

export default function Home(props: Props) {
  return (
    <Layout>
      <div className="bg-blue-500">
        <Counter />

        <div>Time: {props.time}</div>
        <div>Route: {props.route}</div>

        <a href="/about">About</a>
      </div>
    </Layout>
  );
}
