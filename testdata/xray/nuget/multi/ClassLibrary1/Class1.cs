using System;
using System.Security.AccessControl;

namespace ClassLibrary1
{
  public class Class1
  {
    public void foo()
    {
      var foo = AccessControlType.Allow;
      Console.Error.WriteLine(foo.ToString());
    }
    
  }
}
