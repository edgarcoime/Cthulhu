"use server";

import Link from "next/link";
import FileDropzone from "./FileDropzone"

export default async function Home() {


  return (
    <div className="h-screen w-screen flex flex-col justify-center">
      <section className="flex flex-col gap-6">
        <div>
          <h1 className="font-black text-6xl text-center">CTHULHU</h1>
        </div>

        <div className="flex justify-center">
          <FileDropzone />
        </div>
      </section>
    </div>
  );
}
