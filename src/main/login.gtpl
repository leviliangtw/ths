<html>
	<head>
		<title>Texas Poker Online!!!</title>
	</head>

	<style type='text/css'>
		fieldset {
			background-color:#dddddd
			margin:10px;
			padding:10px;
			width:900px;
			line-height:2;
		}
	</style>

	<body style="background-color:#666666">
		<header style="margin:30px;">
			<center><h1 style="color:#FFFFFF;font-family:verdana;">Texas Poker Online</h1></center>
		</header>

	<form action="GameStart" method="post"><center>
		<div style="background-color:#363636;width:900px;height:330px">
			<div style="text-align:center;float:left;width:440px;margin-left:40px;margin-top:30px;">
				<p style="text-align:left;size:50px;font-size:24px;color:#FFFFFF;font-family:verdana;">Sign In to Texas Poker</p>
				<fieldset {{.FieldsetDisable}} style=";color:#FFFFFF;text-align:left;border-color:#CCCCCC;border-style:solid;font-family:verdana;font-size:14px;width:390px;">
				<legend style="font-family:verdana;" ><strong>TEAM</strong></legend>
					<label><input type="radio" name="team" value="1">Team 1</label>
					<label><input type="radio" name="team" value="2">Team 2</label>
					<label><input type="radio" name="team" value="3">Team 3</label>
					<label><input type="radio" name="team" value="4">Team 4</label>
					<label><input type="radio" name="team" value="5">Team 5</label>
				</fieldset>
				<br/><br/>
				<div style="text-align:left;float:left;width:200px;">
					<input type="submit" method="post" name="SignIn" value="Sign In" style="display:{{.SignInDisplay}};height:40px;width:100px;font-size:14px;background-color:#bfffdf;border-style:none;font-family:Microsoft JhengHei;">
					<input type="submit" method="post" name="SignOut" value="Sign Out" style="display:{{.SignOutDisplay}};height:40px;width:100px;font-size:14px;background-color:#bfffdf;border-style:none;font-family:Microsoft JhengHei;">
				</div>
				<div style="text-align:left;float:right;width:200px;">
					<input type="submit" method="post" name="Shuffle" value="Shuffle" style="display:{{.StartGameDisplay}};height:40px;width:100px;font-size:14px;background-color:#bfffdf;border-style:none;font-family:Microsoft JhengHei;">
					<input type="hidden" value="ExitGame" style="display:{{.ExitGameDisplay}};height:40px;width:100px;font-size:14px;background-color:#bfffdf;border-style:none;font-family:Microsoft JhengHei;">
				</div>
			</div>
			<div style="text-align:left;float:right;width:380px;margin-right:40px;margin-top:30px;color:#FFFFFF;font-family:Microsoft JhengHei;">
				<p style="text-align:left;size:50px;font-size:18px;color:#FFFFFF;font-family:verdana;">Notice:</p>
				<ul>
				<li>Select your team number correctly.</li>
				<li>Click the botton "Sign In" to get a ID.</li>
				<li>When every teams have signed in, the game will start.</li>
				<li>We will post two kinds of data to you:</li>
					<ul>
					<li>Info</li>
					<li>XmlToClient</li>
					</ul>
				<li>And you should post back "XmlToServer".</li>
				</ul>
				<p style="display:{{.IsDisplay}};font-size:18px;font-family:verdana;">Your ID:{{.ID}}</p>
			</div>
			<br/>
		</div>
		<input type="hidden" value="{{.GameStartDisplay}}">
	</center></form>
	</body>
</html>
