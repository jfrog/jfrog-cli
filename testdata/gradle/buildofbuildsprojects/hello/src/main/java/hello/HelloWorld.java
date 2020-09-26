package hello;


import org.jfrog.cli.greeter.Greeter;
public class HelloWorld {
  public static void main(String[] args) {

	Greeter greeter = new Greeter();
	System.out.println(greeter.sayHello());
  }
}